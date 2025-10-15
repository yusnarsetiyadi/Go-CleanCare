package work

import (
	"bytes"
	"cleancare/internal/abstraction"
	"cleancare/internal/dto"
	"cleancare/internal/factory"
	"cleancare/internal/model"
	"cleancare/internal/repository"
	"cleancare/pkg/constant"
	"cleancare/pkg/gdrive"
	"cleancare/pkg/util/general"
	"cleancare/pkg/util/response"
	"cleancare/pkg/util/trxmanager"
	"fmt"
	"net/http"
	"slices"
	"strconv"
	"strings"

	"github.com/go-redis/redis/v8"
	"github.com/jung-kurt/gofpdf"
	"github.com/pkg/errors"
	"github.com/sirupsen/logrus"
	"github.com/xuri/excelize/v2"
	"google.golang.org/api/drive/v3"
	"gorm.io/gorm"
)

type Service interface {
	Create(ctx *abstraction.Context, payload *dto.WorkCreateRequest) (map[string]interface{}, error)
	Delete(ctx *abstraction.Context, payload *dto.WorkDeleteByIDRequest) (map[string]interface{}, error)
	Find(ctx *abstraction.Context) (map[string]interface{}, error)
	FindById(ctx *abstraction.Context, payload *dto.WorkFindByIDRequest) (map[string]interface{}, error)
	Update(ctx *abstraction.Context, payload *dto.WorkUpdateRequest) (map[string]interface{}, error)
	Export(ctx *abstraction.Context, payload *dto.WorkExportRequest) (string, *bytes.Buffer, string, error)
	DashboardAdmin(ctx *abstraction.Context, payload *dto.WorkDashboardAdminRequest) (map[string]interface{}, error)
	DashboardStaf(ctx *abstraction.Context, payload *dto.WorkDashboardStafRequest) (map[string]interface{}, error)
}

type service struct {
	TaskRepository     repository.Task
	TaskTypeRepository repository.TaskType
	WorkRepository     repository.Work
	CommentReposiory   repository.Comment

	DB      *gorm.DB
	DbRedis *redis.Client
	sDrive  *drive.Service
	fDrive  *drive.File
}

func NewService(f *factory.Factory) Service {
	return &service{
		TaskRepository:     f.TaskRepository,
		TaskTypeRepository: f.TaskTypeRepository,
		WorkRepository:     f.WorkRepository,
		CommentReposiory:   f.CommentRepository,

		DB:      f.Db,
		DbRedis: f.DbRedis,
		sDrive:  f.GDrive.Service,
		fDrive:  f.GDrive.FolderCleanCare,
	}
}

func (s *service) Create(ctx *abstraction.Context, payload *dto.WorkCreateRequest) (map[string]interface{}, error) {
	var (
		allFileUploaded []string = nil
		imageBefore     *string
		imageAfter      *string
	)
	if err := trxmanager.New(s.DB).WithTrx(ctx, func(ctx *abstraction.Context) error {
		if ctx.Auth.RoleID != constant.ROLE_ID_STAFF {
			return response.ErrorBuilder(http.StatusBadRequest, errors.New("bad_request"), "this role is not permitted")
		}

		taskData, err := s.TaskRepository.FindById(ctx, payload.TaskId)
		if err != nil && err.Error() != "record not found" {
			return response.ErrorBuilder(http.StatusInternalServerError, err, "server_error")
		}
		if taskData == nil {
			return response.ErrorBuilder(http.StatusBadRequest, errors.New("bad_request"), "task not found")
		}

		taskTypeData, err := s.TaskTypeRepository.FindById(ctx, payload.TaskTypeId)
		if err != nil && err.Error() != "record not found" {
			return response.ErrorBuilder(http.StatusInternalServerError, err, "server_error")
		}
		if taskTypeData == nil {
			return response.ErrorBuilder(http.StatusBadRequest, errors.New("bad_request"), "task type not found")
		}

		if payload.ImageBefore != nil {
			file := payload.ImageBefore[0]

			f, err := file.Open()
			if err != nil {
				return response.ErrorBuilder(http.StatusInternalServerError, err, "server_error")
			}
			defer f.Close()

			isImageFile, fullFileName := general.ValidateImage(file.Filename)
			if !isImageFile {
				return response.ErrorBuilder(http.StatusBadRequest, errors.New("bad_request"), fmt.Sprintf("file format for %s is not approved", file.Filename))
			}

			newFile, err := gdrive.CreateFile(s.sDrive, fullFileName, "application/octet-stream", f, s.fDrive.Id)
			if err != nil {
				return response.ErrorBuilder(http.StatusInternalServerError, err, "server_error")
			}
			allFileUploaded = append(allFileUploaded, newFile.Id)

			imgFileDelimiter := general.JoinFileAndNameWithDelimiter(newFile.Id, newFile.Name)
			imageBefore = &imgFileDelimiter
		}

		if payload.ImageAfter != nil {
			file := payload.ImageAfter[0]

			f, err := file.Open()
			if err != nil {
				return response.ErrorBuilder(http.StatusInternalServerError, err, "server_error")
			}
			defer f.Close()

			isImageFile, fullFileName := general.ValidateImage(file.Filename)
			if !isImageFile {
				return response.ErrorBuilder(http.StatusBadRequest, errors.New("bad_request"), fmt.Sprintf("file format for %s is not approved", file.Filename))
			}

			newFile, err := gdrive.CreateFile(s.sDrive, fullFileName, "application/octet-stream", f, s.fDrive.Id)
			if err != nil {
				return response.ErrorBuilder(http.StatusInternalServerError, err, "server_error")
			}
			allFileUploaded = append(allFileUploaded, newFile.Id)

			imgFileDelimiter := general.JoinFileAndNameWithDelimiter(newFile.Id, newFile.Name)
			imageAfter = &imgFileDelimiter
		}

		modelWork := &model.WorkEntityModel{
			Context: ctx,
			WorkEntity: model.WorkEntity{
				UserId:      ctx.Auth.ID,
				TaskId:      payload.TaskId,
				TaskTypeId:  payload.TaskTypeId,
				Floor:       payload.Floor,
				Info:        payload.Info,
				ImageBefore: imageBefore,
				ImageAfter:  imageAfter,
				IsDelete:    false,
			},
		}
		if err = s.WorkRepository.Create(ctx, modelWork).Error; err != nil {
			return response.ErrorBuilder(http.StatusInternalServerError, err, "server_error")
		}

		return nil
	}); err != nil {
		for _, v := range allFileUploaded {
			errDel := gdrive.DeleteFile(s.sDrive, v)
			if errDel != nil {
				logrus.Error("error delete file for error trxmanager:", errDel.Error())
			}
		}
		return nil, err
	}
	return map[string]interface{}{
		"message": "success create!",
	}, nil
}

func (s *service) Delete(ctx *abstraction.Context, payload *dto.WorkDeleteByIDRequest) (map[string]interface{}, error) {
	if err := trxmanager.New(s.DB).WithTrx(ctx, func(ctx *abstraction.Context) error {
		workData, err := s.WorkRepository.FindById(ctx, payload.ID)
		if err != nil && err.Error() != "record not found" {
			return response.ErrorBuilder(http.StatusInternalServerError, err, "server_error")
		}
		if workData == nil {
			return response.ErrorBuilder(http.StatusBadRequest, errors.New("bad_request"), "work not found")
		}

		if ctx.Auth.ID != workData.UserId {
			return response.ErrorBuilder(http.StatusBadRequest, errors.New("bad_request"), "this role is not permitted")
		}

		newWorkData := new(model.WorkEntityModel)
		newWorkData.Context = ctx
		newWorkData.ID = workData.ID
		newWorkData.IsDelete = true

		if err = s.WorkRepository.Update(ctx, newWorkData).Error; err != nil {
			return response.ErrorBuilder(http.StatusInternalServerError, err, "server_error")
		}

		return nil
	}); err != nil {
		return nil, err
	}
	return map[string]interface{}{
		"message": "success delete!",
	}, nil
}

func (s *service) Find(ctx *abstraction.Context) (map[string]interface{}, error) {
	var res []map[string]interface{} = nil
	data, err := s.WorkRepository.Find(ctx, false)
	if err != nil && err.Error() != "record not found" {
		return nil, response.ErrorBuilder(http.StatusInternalServerError, err, "server_error")
	}
	count, err := s.WorkRepository.Count(ctx)
	if err != nil && err.Error() != "record not found" {
		return nil, response.ErrorBuilder(http.StatusInternalServerError, err, "server_error")
	}
	for _, v := range data {
		// check comment unread
		hasUnreadComment := false
		commentData, err := s.CommentReposiory.FindByWorkIdArr(ctx, v.ID, true)
		if err != nil && err.Error() != "record not found" {
			return nil, response.ErrorBuilder(http.StatusInternalServerError, err, "server_error")
		}
		for _, comment := range commentData {
			userDataUnreadComment := general.GetUserIdArrayFromKeyRedis(s.DbRedis, general.GenerateRedisKeyUnreadComment(comment.ID))
			if slices.Contains(userDataUnreadComment, strconv.Itoa(ctx.Auth.ID)) {
				hasUnreadComment = true
			}
		}

		// from user id
		resUser := map[string]interface{}{
			"id":           v.User.ID,
			"name":         v.User.Name,
			"profile":      v.User.Profile,
			"profile_name": v.User.ProfileName,
		}
		if v.User.Profile != nil {
			profile, err := gdrive.GetFile(s.sDrive, *v.User.Profile)
			if err != nil {
				return nil, response.ErrorBuilder(http.StatusBadRequest, errors.New("bad_request"), "profile not found")
			}
			resUser["profile"] = map[string]interface{}{
				// "view_saved": general.ConvertLinkToFileSaved(profile.WebContentLink, profile.Name, profile.FileExtension),
				"view":    "https://lh3.googleusercontent.com/d/" + *v.User.Profile,
				"content": profile.WebContentLink,
				"ext":     profile.FileExtension,
				"name":    profile.Name,
				"id":      profile.Id,
			}
		}

		// work data
		resData := map[string]interface{}{
			"id":   v.ID,
			"user": resUser,
			"task": map[string]interface{}{
				"id":   v.Task.ID,
				"name": v.Task.Name,
			},
			"task_type": map[string]interface{}{
				"id":   v.TaskType.ID,
				"name": v.TaskType.Name,
			},
			"floor":          v.Floor,
			"info":           v.Info,
			"unread_comment": hasUnreadComment,
			"created_at":     general.FormatWithZWithoutChangingTime(v.CreatedAt),
			"updated_at":     general.FormatWithZWithoutChangingTime(*v.UpdatedAt),
		}

		res = append(res, resData)
	}
	return map[string]interface{}{
		"count": count,
		"data":  res,
	}, nil
}

func (s *service) FindById(ctx *abstraction.Context, payload *dto.WorkFindByIDRequest) (map[string]interface{}, error) {
	var res map[string]interface{} = nil
	data, err := s.WorkRepository.FindById(ctx, payload.ID)
	if err != nil && err.Error() != "record not found" {
		return nil, response.ErrorBuilder(http.StatusInternalServerError, err, "server_error")
	}
	if data != nil {
		// from user id
		resUser := map[string]interface{}{
			"id":           data.User.ID,
			"name":         data.User.Name,
			"profile":      data.User.Profile,
			"profile_name": data.User.ProfileName,
		}
		if data.User.Profile != nil {
			profile, err := gdrive.GetFile(s.sDrive, *data.User.Profile)
			if err != nil {
				return nil, response.ErrorBuilder(http.StatusBadRequest, errors.New("bad_request"), "profile not found")
			}
			resUser["profile"] = map[string]interface{}{
				// "view_saved": general.ConvertLinkToFileSaved(profile.WebContentLink, profile.Name, profile.FileExtension),
				"view":    "https://lh3.googleusercontent.com/d/" + *data.User.Profile,
				"content": profile.WebContentLink,
				"ext":     profile.FileExtension,
				"name":    profile.Name,
				"id":      profile.Id,
			}
		}

		res = map[string]interface{}{
			"id":   data.ID,
			"user": resUser,
			"task": map[string]interface{}{
				"id":   data.Task.ID,
				"name": data.Task.Name,
			},
			"task_type": map[string]interface{}{
				"id":   data.TaskType.ID,
				"name": data.TaskType.Name,
			},
			"floor":        data.Floor,
			"info":         data.Info,
			"image_before": data.ImageBefore,
			"image_after":  data.ImageAfter,
			"created_at":   general.FormatWithZWithoutChangingTime(data.CreatedAt),
			"updated_at":   general.FormatWithZWithoutChangingTime(*data.UpdatedAt),
		}
		if data.ImageBefore != nil {
			image_before, err := gdrive.GetFile(s.sDrive, *data.ImageBefore)
			if err != nil {
				return nil, response.ErrorBuilder(http.StatusBadRequest, errors.New("bad_request"), "image_before not found")
			}
			res["image_before"] = map[string]interface{}{
				// "view_saved": general.ConvertLinkToFileSaved(image_before.WebContentLink, image_before.Name, image_before.FileExtension),
				"view":    "https://lh3.googleusercontent.com/d/" + *data.ImageBefore,
				"content": image_before.WebContentLink,
				"ext":     image_before.FileExtension,
				"name":    image_before.Name,
				"id":      image_before.Id,
			}
		}
		if data.ImageAfter != nil {
			image_after, err := gdrive.GetFile(s.sDrive, *data.ImageAfter)
			if err != nil {
				return nil, response.ErrorBuilder(http.StatusBadRequest, errors.New("bad_request"), "image_after not found")
			}
			res["image_after"] = map[string]interface{}{
				// "view_saved": general.ConvertLinkToFileSaved(image_after.WebContentLink, image_after.Name, image_after.FileExtension),
				"view":    "https://lh3.googleusercontent.com/d/" + *data.ImageAfter,
				"content": image_after.WebContentLink,
				"ext":     image_after.FileExtension,
				"name":    image_after.Name,
				"id":      image_after.Id,
			}
		}
	}
	return map[string]interface{}{
		"data": res,
	}, nil
}

func (s *service) Update(ctx *abstraction.Context, payload *dto.WorkUpdateRequest) (map[string]interface{}, error) {
	var (
		allFileUploaded []string
		allFileOld      []string
	)
	if err := trxmanager.New(s.DB).WithTrx(ctx, func(ctx *abstraction.Context) error {
		workData, err := s.WorkRepository.FindById(ctx, payload.ID)
		if err != nil && err.Error() != "record not found" {
			return response.ErrorBuilder(http.StatusInternalServerError, err, "server_error")
		}
		if workData == nil {
			return response.ErrorBuilder(http.StatusBadRequest, errors.New("bad_request"), "work not found")
		}

		if ctx.Auth.ID != workData.UserId {
			return response.ErrorBuilder(http.StatusBadRequest, errors.New("bad_request"), "this role is not permitted")
		}

		newWorkData := new(model.WorkEntityModel)
		newWorkData.Context = ctx
		newWorkData.ID = payload.ID
		if payload.TaskId != nil {
			taskData, err := s.TaskRepository.FindById(ctx, *payload.TaskId)
			if err != nil && err.Error() != "record not found" {
				return response.ErrorBuilder(http.StatusInternalServerError, err, "server_error")
			}
			if taskData == nil {
				return response.ErrorBuilder(http.StatusBadRequest, errors.New("bad_request"), "task id not found")
			}
			newWorkData.TaskId = *payload.TaskId
		}
		if payload.TaskTypeId != nil {
			taskTypeData, err := s.TaskTypeRepository.FindById(ctx, *payload.TaskTypeId)
			if err != nil && err.Error() != "record not found" {
				return response.ErrorBuilder(http.StatusInternalServerError, err, "server_error")
			}
			if taskTypeData == nil {
				return response.ErrorBuilder(http.StatusBadRequest, errors.New("bad_request"), "task type id not found")
			}
			newWorkData.TaskTypeId = *payload.TaskTypeId
		}
		if payload.Floor != nil {
			newWorkData.Floor = *payload.Floor
		}
		if payload.Info != nil {
			newWorkData.Info = *payload.Info
		}
		if payload.ImageBefore != nil {
			file := payload.ImageBefore[0]

			f, err := file.Open()
			if err != nil {
				return response.ErrorBuilder(http.StatusInternalServerError, err, "server_error")
			}
			defer f.Close()

			isImageFile, fullFileName := general.ValidateImage(file.Filename)
			if !isImageFile {
				return response.ErrorBuilder(http.StatusBadRequest, errors.New("bad_request"), fmt.Sprintf("file format for %s is not approved", file.Filename))
			}

			newFile, err := gdrive.CreateFile(s.sDrive, fullFileName, "application/octet-stream", f, s.fDrive.Id)
			if err != nil {
				return response.ErrorBuilder(http.StatusInternalServerError, err, "server_error")
			}
			allFileUploaded = append(allFileUploaded, newFile.Id)

			imgFileDelimiter := general.JoinFileAndNameWithDelimiter(newFile.Id, newFile.Name)
			newWorkData.ImageBefore = &imgFileDelimiter

			if workData.ImageBefore != nil {
				imageBeforeFile, _ := general.SplitFileAndNameWithDelimiter(*workData.ImageBefore)
				allFileOld = append(allFileOld, imageBeforeFile)
			}
		} else {
			if payload.DeleteImageBefore != nil && *payload.DeleteImageBefore {
				imageBeforeFile, _ := general.SplitFileAndNameWithDelimiter(*workData.ImageBefore)
				errDel := gdrive.DeleteFile(s.sDrive, imageBeforeFile)
				if errDel != nil {
					logrus.Error("error delete file for cover:", errDel.Error())
				}
				if err = s.WorkRepository.UpdateToNull(ctx, newWorkData, "image_before").Error; err != nil {
					return response.ErrorBuilder(http.StatusInternalServerError, err, "server_error")
				}
			}
		}
		if payload.ImageAfter != nil {
			file := payload.ImageAfter[0]

			f, err := file.Open()
			if err != nil {
				return response.ErrorBuilder(http.StatusInternalServerError, err, "server_error")
			}
			defer f.Close()

			isImageFile, fullFileName := general.ValidateImage(file.Filename)
			if !isImageFile {
				return response.ErrorBuilder(http.StatusBadRequest, errors.New("bad_request"), fmt.Sprintf("file format for %s is not approved", file.Filename))
			}

			newFile, err := gdrive.CreateFile(s.sDrive, fullFileName, "application/octet-stream", f, s.fDrive.Id)
			if err != nil {
				return response.ErrorBuilder(http.StatusInternalServerError, err, "server_error")
			}
			allFileUploaded = append(allFileUploaded, newFile.Id)

			imgFileDelimiter := general.JoinFileAndNameWithDelimiter(newFile.Id, newFile.Name)
			newWorkData.ImageAfter = &imgFileDelimiter

			if workData.ImageAfter != nil {
				imageAfterFile, _ := general.SplitFileAndNameWithDelimiter(*workData.ImageAfter)
				allFileOld = append(allFileOld, imageAfterFile)
			}
		} else {
			if payload.DeleteImageAfter != nil && *payload.DeleteImageAfter {
				imageAfterFile, _ := general.SplitFileAndNameWithDelimiter(*workData.ImageAfter)
				errDel := gdrive.DeleteFile(s.sDrive, imageAfterFile)
				if errDel != nil {
					logrus.Error("error delete file for cover:", errDel.Error())
				}
				if err = s.WorkRepository.UpdateToNull(ctx, newWorkData, "image_after").Error; err != nil {
					return response.ErrorBuilder(http.StatusInternalServerError, err, "server_error")
				}
			}
		}
		if err = s.WorkRepository.Update(ctx, newWorkData).Error; err != nil {
			return response.ErrorBuilder(http.StatusInternalServerError, err, "server_error")
		}

		return nil
	}); err != nil {
		for _, v := range allFileUploaded {
			errDel := gdrive.DeleteFile(s.sDrive, v)
			if errDel != nil {
				logrus.Error("error delete file for error trxmanager:", errDel.Error())
			}
		}
		return nil, err
	}

	for _, v := range allFileOld {
		errDel := gdrive.DeleteFile(s.sDrive, v)
		if errDel != nil {
			logrus.Error("error delete file old after trxmanager:", errDel.Error())
		}
	}

	return map[string]interface{}{
		"message": "success update!",
	}, nil
}

func (s *service) Export(ctx *abstraction.Context, payload *dto.WorkExportRequest) (string, *bytes.Buffer, string, error) {
	data, err := s.WorkRepository.Find(ctx, true)
	if err != nil && err.Error() != "record not found" {
		return "", nil, "", response.ErrorBuilder(http.StatusInternalServerError, err, "server_error")
	}

	reportDate := ""
	if ctx.QueryParam("created_at") != "" {
		val := general.SanitizeStringDateBetween(ctx.QueryParam("created_at"))
		valDate := strings.Split(val, "_")
		reportDate = valDate[0]
	}

	if payload.Format == "pdf" {
		pdf := gofpdf.New("L", "mm", "A4", "")
		pdf.SetMargins(10, 10, 10)
		pdf.AddPage()

		pdf.SetFont("Arial", "B", 16)
		pdf.Cell(0, 10, fmt.Sprintf(
			"CleanCare - Laporan Pekerjaan Petugas Kebersihan (%s)",
			general.ConvertDateToIndonesian(reportDate),
		))
		pdf.Ln(12)
		pdf.SetFont("Arial", "B", 10)
		header := []string{
			"No", "Petugas Kebersihan", "Pekerjaan", "Jenis Pekerjaan",
			"Lantai", "Keterangan", "Sebelum", "Sesudah", "Tanggal",
		}
		colWidths := []float64{
			10, 38, 30, 30, 18, 52, 30, 30, 39,
		}
		xStart := pdf.GetX()
		yStart := pdf.GetY()
		headerHeight := 8.0

		for i, str := range header {
			pdf.Rect(xStart, yStart, colWidths[i], headerHeight, "D")
			pdf.MultiCell(colWidths[i], 5, str, "", "C", false)
			xStart += colWidths[i]
			pdf.SetXY(xStart, yStart)
		}
		pdf.Ln(headerHeight)
		pdf.SetFont("Arial", "", 9)

		for i, v := range data {
			no := fmt.Sprintf("%d", i+1)
			linkImageBefore := ""
			linkImageAfter := ""

			if v.ImageBefore != nil {
				linkImageBefore = "https://lh3.googleusercontent.com/d/" + *v.ImageBefore
			}
			if v.ImageAfter != nil {
				linkImageAfter = "https://lh3.googleusercontent.com/d/" + *v.ImageAfter
			}

			row := []string{
				no,
				v.User.Name,
				v.Task.Name,
				v.TaskType.Name,
				v.Floor,
				v.Info,
				linkImageBefore,
				linkImageAfter,
				general.ConvertDateTimeToIndonesian(
					v.CreatedAt.Format("2006-01-02 15:04:05"),
				),
			}

			startX := pdf.GetX()
			startY := pdf.GetY()
			maxHeight := 0.0
			for j, txt := range row {
				lines := pdf.SplitLines([]byte(txt), colWidths[j])
				h := float64(len(lines)) * 5
				if h > maxHeight {
					maxHeight = h
				}
			}
			x := startX
			for j, txt := range row {
				y := pdf.GetY()
				pdf.Rect(x, y, colWidths[j], maxHeight, "D")
				pdf.MultiCell(colWidths[j], 5, txt, "", "", false)
				x += colWidths[j]
				pdf.SetXY(x, y)
			}
			pdf.SetXY(startX, startY+maxHeight)
		}
		var buf bytes.Buffer
		if err := pdf.Output(&buf); err != nil {
			return "", nil, "", response.ErrorBuilder(http.StatusInternalServerError, err, "server_error")
		}

		filename := fmt.Sprintf("(%s) CleanCare - Laporan Pekerjaan Petugas Kebersihan.pdf", strings.ReplaceAll(reportDate, "-", ""))
		return filename, &buf, "pdf", nil

	} else {
		f := excelize.NewFile()
		sheet := "CleanCare"
		index, err := f.NewSheet(general.TruncateSheetName(sheet))
		if err != nil {
			return "", nil, "", response.ErrorBuilder(http.StatusInternalServerError, err, "server_error")
		}
		f.DeleteSheet("Sheet1")
		f.SetActiveSheet(index)
		f.SetCellValue(sheet, "A1", "No")
		f.SetCellValue(sheet, "B1", "Petugas Kebersihan")
		f.SetCellValue(sheet, "C1", "Pekerjaan")
		f.SetCellValue(sheet, "D1", "Jenis Pekerjaan")
		f.SetCellValue(sheet, "E1", "Lantai")
		f.SetCellValue(sheet, "F1", "Keterangan")
		f.SetCellValue(sheet, "G1", "Sebelum")
		f.SetCellValue(sheet, "H1", "Sesudah")
		f.SetCellValue(sheet, "I1", "Tanggal")
		for i, v := range data {
			colA := fmt.Sprintf("A%d", i+2)
			colB := fmt.Sprintf("B%d", i+2)
			colC := fmt.Sprintf("C%d", i+2)
			colD := fmt.Sprintf("D%d", i+2)
			colE := fmt.Sprintf("E%d", i+2)
			colF := fmt.Sprintf("F%d", i+2)
			colG := fmt.Sprintf("G%d", i+2)
			colH := fmt.Sprintf("H%d", i+2)
			colI := fmt.Sprintf("I%d", i+2)
			var (
				linkImageBefore string
				linkImageAfter  string
			)
			if v.ImageBefore != nil {
				linkImageBefore = "https://lh3.googleusercontent.com/d/" + *v.ImageBefore
			}
			if v.ImageAfter != nil {
				linkImageAfter = "https://lh3.googleusercontent.com/d/" + *v.ImageAfter
			}
			no := i + 1
			f.SetCellValue(sheet, colA, no)
			f.SetCellValue(sheet, colB, v.User.Name)
			f.SetCellValue(sheet, colC, v.Task.Name)
			f.SetCellValue(sheet, colD, v.TaskType.Name)
			f.SetCellValue(sheet, colE, v.Floor)
			f.SetCellValue(sheet, colF, v.Info)
			f.SetCellValue(sheet, colG, linkImageBefore)
			f.SetCellValue(sheet, colH, linkImageAfter)
			f.SetCellValue(sheet, colI, general.ConvertDateTimeToIndonesian(v.CreatedAt.Format("2006-01-02 15:04:05")))
		}

		var buf bytes.Buffer
		if err := f.Write(&buf); err != nil {
			return "", nil, "", response.ErrorBuilder(http.StatusInternalServerError, err, "server_error")
		}
		filename := fmt.Sprintf("(%s) CleanCare - Laporan Pekerjaan Petugas Kebersihan.xlsx", strings.ReplaceAll(reportDate, "-", ""))
		return filename, &buf, "excel", nil
	}
}

func (s *service) DashboardAdmin(ctx *abstraction.Context, payload *dto.WorkDashboardAdminRequest) (map[string]interface{}, error) {
	floorSummary, userSummary, errFloor, errUser := s.WorkRepository.FindByTaskIdArrAdmin(ctx, payload.TaskId, payload.CreatedAt, true)
	if errFloor != nil && errFloor.Error() != "record not found" {
		return nil, response.ErrorBuilder(http.StatusInternalServerError, errFloor, "server_error")
	}
	if errUser != nil && errUser.Error() != "record not found" {
		return nil, response.ErrorBuilder(http.StatusInternalServerError, errUser, "server_error")
	}
	if payload.TaskId == constant.TASK_ID_DAILY {
		return map[string]interface{}{
			"data": floorSummary,
		}, nil
	} else {
		return map[string]interface{}{
			"data": userSummary,
		}, nil
	}
}

func (s *service) DashboardStaf(ctx *abstraction.Context, payload *dto.WorkDashboardStafRequest) (map[string]interface{}, error) {
	taskTypeSummary, err := s.WorkRepository.FindByTaskIdArrStaf(ctx, payload.TaskId, payload.CreatedAt, true)
	if err != nil && err.Error() != "record not found" {
		return nil, response.ErrorBuilder(http.StatusInternalServerError, err, "server_error")
	}
	return map[string]interface{}{
		"data": taskTypeSummary,
	}, nil
}
