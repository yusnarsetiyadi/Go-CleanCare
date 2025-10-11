package user

import (
	"bytes"
	"fmt"
	"iss_cleancare/internal/abstraction"
	"iss_cleancare/internal/dto"
	"iss_cleancare/internal/factory"
	"iss_cleancare/internal/model"
	"iss_cleancare/internal/repository"
	"iss_cleancare/pkg/constant"
	"iss_cleancare/pkg/gdrive"
	"iss_cleancare/pkg/util/general"
	"iss_cleancare/pkg/util/response"
	"iss_cleancare/pkg/util/trxmanager"
	"net/http"

	"github.com/go-redis/redis/v8"
	"github.com/jung-kurt/gofpdf"
	"github.com/pkg/errors"
	"github.com/sirupsen/logrus"
	"github.com/xuri/excelize/v2"
	"golang.org/x/crypto/bcrypt"
	"google.golang.org/api/drive/v3"
	"gorm.io/gorm"
)

type Service interface {
	Create(ctx *abstraction.Context, payload *dto.UserCreateRequest) (map[string]interface{}, error)
	Find(ctx *abstraction.Context) (map[string]interface{}, error)
	FindById(ctx *abstraction.Context, payload *dto.UserFindByIDRequest) (map[string]interface{}, error)
	Update(ctx *abstraction.Context, payload *dto.UserUpdateRequest) (map[string]interface{}, error)
	Delete(ctx *abstraction.Context, payload *dto.UserDeleteByIDRequest) (map[string]interface{}, error)
	ChangePassword(ctx *abstraction.Context, payload *dto.UserChangePasswordRequest) (map[string]interface{}, error)
	Export(ctx *abstraction.Context, payload *dto.UserExportRequest) (string, *bytes.Buffer, string, error)
}

type service struct {
	UserRepository repository.User
	RoleRepository repository.Role

	DB      *gorm.DB
	DbRedis *redis.Client
	sDrive  *drive.Service
	fDrive  *drive.File
}

func NewService(f *factory.Factory) Service {
	return &service{
		UserRepository: f.UserRepository,
		RoleRepository: f.RoleRepository,

		DB:      f.Db,
		DbRedis: f.DbRedis,
		sDrive:  f.GDrive.Service,
		fDrive:  f.GDrive.FolderIss,
	}
}

func (s *service) Create(ctx *abstraction.Context, payload *dto.UserCreateRequest) (map[string]interface{}, error) {
	if err := trxmanager.New(s.DB).WithTrx(ctx, func(ctx *abstraction.Context) error {
		if ctx.Auth.RoleID != constant.ROLE_ID_ADMIN {
			return response.ErrorBuilder(http.StatusBadRequest, errors.New("bad_request"), "this role is not permitted")
		}

		userNumber, err := s.UserRepository.FindByNumberId(ctx, payload.NumberId)
		if err != nil && err.Error() != "record not found" {
			return response.ErrorBuilder(http.StatusInternalServerError, err, "server_error")
		}
		if userNumber != nil {
			return response.ErrorBuilder(http.StatusBadRequest, errors.New("bad_request"), "number id already exist")
		}

		roleData, err := s.RoleRepository.FindById(ctx, payload.RoleId)
		if err != nil && err.Error() != "record not found" {
			return response.ErrorBuilder(http.StatusInternalServerError, err, "server_error")
		}
		if roleData == nil {
			return response.ErrorBuilder(http.StatusBadRequest, errors.New("bad_request"), "role not found")
		}

		modelUser := &model.UserEntityModel{
			Context: ctx,
			UserEntity: model.UserEntity{
				NumberId: payload.NumberId,
				Name:     payload.Name,
				RoleId:   payload.RoleId,
				IsDelete: false,
			},
		}
		if err = s.UserRepository.Create(ctx, modelUser).Error; err != nil {
			return response.ErrorBuilder(http.StatusInternalServerError, err, "server_error")
		}

		return nil
	}); err != nil {
		return nil, err
	}
	return map[string]interface{}{
		"message": "success create!",
	}, nil
}

func (s *service) Find(ctx *abstraction.Context) (map[string]interface{}, error) {
	var res []map[string]interface{} = nil
	data, err := s.UserRepository.Find(ctx, false)
	if err != nil && err.Error() != "record not found" {
		return nil, response.ErrorBuilder(http.StatusInternalServerError, err, "server_error")
	}
	count, err := s.UserRepository.Count(ctx)
	if err != nil && err.Error() != "record not found" {
		return nil, response.ErrorBuilder(http.StatusInternalServerError, err, "server_error")
	}
	for _, v := range data {
		email := "-"
		verified := false
		if v.Email != nil {
			email = *v.Email
			verified = true
		}
		res = append(res, map[string]interface{}{
			"id":         v.ID,
			"number_id":  v.NumberId,
			"name":       v.Name,
			"email":      email,
			"verified":   verified,
			"is_delete":  v.IsDelete,
			"created_at": general.FormatWithZWithoutChangingTime(v.CreatedAt),
			"updated_at": general.FormatWithZWithoutChangingTime(*v.UpdatedAt),
			"role": map[string]interface{}{
				"id":   v.Role.ID,
				"name": v.Role.Name,
			},
		})
	}
	return map[string]interface{}{
		"count": count,
		"data":  res,
	}, nil
}

func (s *service) FindById(ctx *abstraction.Context, payload *dto.UserFindByIDRequest) (map[string]interface{}, error) {
	var res map[string]interface{} = nil
	data, err := s.UserRepository.FindById(ctx, payload.ID)
	if err != nil && err.Error() != "record not found" {
		return nil, response.ErrorBuilder(http.StatusInternalServerError, err, "server_error")
	}
	if data != nil {
		res = map[string]interface{}{
			"id":           data.ID,
			"number_id":    data.NumberId,
			"name":         data.Name,
			"email":        data.Email,
			"is_delete":    data.IsDelete,
			"created_at":   general.FormatWithZWithoutChangingTime(data.CreatedAt),
			"updated_at":   general.FormatWithZWithoutChangingTime(*data.UpdatedAt),
			"profile":      data.Profile,
			"profile_name": data.ProfileName,
			"role": map[string]interface{}{
				"id":   data.Role.ID,
				"name": data.Role.Name,
			},
		}

		if data.Profile != nil {
			profile, err := gdrive.GetFile(s.sDrive, *data.Profile)
			if err != nil {
				return nil, response.ErrorBuilder(http.StatusBadRequest, errors.New("bad_request"), "profile not found")
			}
			res["profile"] = map[string]interface{}{
				// "view_saved": general.ConvertLinkToFileSaved(profile.WebContentLink, profile.Name, profile.FileExtension),
				"view":    "https://lh3.googleusercontent.com/d/" + *data.Profile,
				"content": profile.WebContentLink,
				"ext":     profile.FileExtension,
				"name":    profile.Name,
				"id":      profile.Id,
			}
		}
	}
	return map[string]interface{}{
		"data": res,
	}, nil
}

func (s *service) Update(ctx *abstraction.Context, payload *dto.UserUpdateRequest) (map[string]interface{}, error) {
	var (
		allFileUploaded []string
		allFileOld      []string
	)
	if err := trxmanager.New(s.DB).WithTrx(ctx, func(ctx *abstraction.Context) error {
		userData, err := s.UserRepository.FindById(ctx, payload.ID)
		if err != nil && err.Error() != "record not found" {
			return response.ErrorBuilder(http.StatusInternalServerError, err, "server_error")
		}
		if userData == nil {
			return response.ErrorBuilder(http.StatusBadRequest, errors.New("bad_request"), "user not found")
		}

		newUserData := new(model.UserEntityModel)
		newUserData.Context = ctx
		newUserData.ID = payload.ID
		if payload.NumberId != nil {
			newUserData.NumberId = *payload.NumberId
		}
		if payload.Name != nil {
			newUserData.Name = *payload.Name
		}
		if payload.Email != nil {
			newUserData.Email = payload.Email
		}
		if payload.RoleId != nil {
			roleData, err := s.RoleRepository.FindById(ctx, *payload.RoleId)
			if err != nil && err.Error() != "record not found" {
				return response.ErrorBuilder(http.StatusInternalServerError, err, "server_error")
			}
			if roleData == nil {
				return response.ErrorBuilder(http.StatusBadRequest, errors.New("bad_request"), "role not found")
			}
			newUserData.RoleId = *payload.RoleId
		}
		if payload.Profile != nil {
			file := payload.Profile[0]

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

			newUserData.Profile = &newFile.Id
			newUserData.ProfileName = &newFile.Name

			if userData.Profile != nil {
				allFileOld = append(allFileOld, *userData.Profile)
			}
		} else {
			if payload.DeleteProfile != nil && *payload.DeleteProfile {
				errDel := gdrive.DeleteFile(s.sDrive, *userData.Profile)
				if errDel != nil {
					logrus.Error("error delete file for cover:", errDel.Error())
				}
				if err = s.UserRepository.UpdateToNull(ctx, newUserData, "profile").Error; err != nil {
					return response.ErrorBuilder(http.StatusInternalServerError, err, "server_error")
				}
				if err = s.UserRepository.UpdateToNull(ctx, newUserData, "profile_name").Error; err != nil {
					return response.ErrorBuilder(http.StatusInternalServerError, err, "server_error")
				}
			}
		}
		if err = s.UserRepository.Update(ctx, newUserData).Error; err != nil {
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

func (s *service) Delete(ctx *abstraction.Context, payload *dto.UserDeleteByIDRequest) (map[string]interface{}, error) {
	if err := trxmanager.New(s.DB).WithTrx(ctx, func(ctx *abstraction.Context) error {
		if ctx.Auth.RoleID != constant.ROLE_ID_ADMIN {
			return response.ErrorBuilder(http.StatusBadRequest, errors.New("bad_request"), "this role is not permitted")
		}

		userData, err := s.UserRepository.FindById(ctx, payload.ID)
		if err != nil && err.Error() != "record not found" {
			return response.ErrorBuilder(http.StatusInternalServerError, err, "server_error")
		}
		if userData == nil {
			return response.ErrorBuilder(http.StatusBadRequest, errors.New("bad_request"), "user not found")
		}

		newUserData := new(model.UserEntityModel)
		newUserData.Context = ctx
		newUserData.ID = userData.ID
		newUserData.IsDelete = true

		if err = s.UserRepository.Update(ctx, newUserData).Error; err != nil {
			return response.ErrorBuilder(http.StatusInternalServerError, err, "server_error")
		}

		userLoginFrom := general.GetRedisUUIDArray(s.DbRedis, general.GenerateRedisKeyUserLogin(userData.ID))
		for _, v := range userLoginFrom {
			general.AppendUUIDToRedisArray(s.DbRedis, constant.REDIS_KEY_AUTO_LOGOUT, v)
		}

		return nil
	}); err != nil {
		return nil, err
	}
	return map[string]interface{}{
		"message": "success delete!",
	}, nil
}

func (s *service) ChangePassword(ctx *abstraction.Context, payload *dto.UserChangePasswordRequest) (map[string]interface{}, error) {
	if err := trxmanager.New(s.DB).WithTrx(ctx, func(ctx *abstraction.Context) error {
		if ctx.Auth.ID != payload.ID {
			return response.ErrorBuilder(http.StatusBadRequest, errors.New("bad_request"), "this user is not permitted")
		}

		userData, err := s.UserRepository.FindById(ctx, payload.ID)
		if err != nil && err.Error() != "record not found" {
			return response.ErrorBuilder(http.StatusInternalServerError, err, "server_error")
		}
		if userData == nil {
			return response.ErrorBuilder(http.StatusBadRequest, errors.New("bad_request"), "user not found")
		}

		if err = bcrypt.CompareHashAndPassword([]byte(*userData.Password), []byte(payload.OldPassword)); err != nil {
			return response.ErrorBuilder(http.StatusBadRequest, err, "old password is wrong")
		}

		if payload.OldPassword == payload.NewPassword {
			return response.ErrorBuilder(http.StatusBadRequest, errors.New("bad_request"), "the new password cannot be the same as the old password")
		}

		hashedPassword, err := bcrypt.GenerateFromPassword([]byte(payload.NewPassword), bcrypt.DefaultCost)
		if err != nil {
			return response.ErrorBuilder(http.StatusInternalServerError, err, "server_error")
		}

		hashPasswordStr := string(hashedPassword)
		newUserData := new(model.UserEntityModel)
		newUserData.Context = ctx
		newUserData.ID = userData.ID
		newUserData.Password = &hashPasswordStr

		if err = s.UserRepository.Update(ctx, newUserData).Error; err != nil {
			return response.ErrorBuilder(http.StatusInternalServerError, err, "server_error")
		}

		userLoginFrom := general.GetRedisUUIDArray(s.DbRedis, general.GenerateRedisKeyUserLogin(userData.ID))
		for _, v := range userLoginFrom {
			general.AppendUUIDToRedisArray(s.DbRedis, constant.REDIS_KEY_AUTO_LOGOUT, v)
		}

		return nil
	}); err != nil {
		return nil, err
	}

	return map[string]interface{}{
		"message": "success change password!",
	}, nil
}

func (s *service) Export(ctx *abstraction.Context, payload *dto.UserExportRequest) (string, *bytes.Buffer, string, error) {
	data, err := s.UserRepository.Find(ctx, true)
	if err != nil && err.Error() != "record not found" {
		return "", nil, "", response.ErrorBuilder(http.StatusInternalServerError, err, "server_error")
	}

	if payload.Format == "pdf" {
		pdf := gofpdf.New("L", "mm", "A4", "")
		pdf.SetMargins(10, 10, 10)
		pdf.AddPage()
		pdf.SetFont("Arial", "B", 16)
		pdf.Cell(0, 10, "ISS CleanCare - Laporan Data Petugas Kebersihan")
		pdf.Ln(12)
		pdf.SetFont("Arial", "B", 10)
		header := []string{
			"No", "Nomor ID", "Nama", "Email",
			"Jabatan", "Tanggal Terdaftar", "Status Verifikasi",
		}
		colWidths := []float64{10, 35, 40, 50, 40, 60, 42}
		for i, str := range header {
			pdf.CellFormat(colWidths[i], 8, str, "1", 0, "C", false, 0, "")
		}
		pdf.Ln(-1)
		pdf.SetFont("Arial", "", 9)

		for i, v := range data {
			no := fmt.Sprintf("%d", i+1)
			email := "-"
			role := ""
			verified := "Belum"

			if v.Email != nil {
				email = *v.Email
				verified = "Sudah"
			}

			if v.RoleId == constant.ROLE_ID_STAFF {
				role = "Petugas Kebersihan"
			} else {
				role = "Supervisor"
			}

			row := []string{
				no,
				v.NumberId,
				v.Name,
				email,
				role,
				general.ConvertDateTimeToIndonesian(v.CreatedAt.Format("2006-01-02 15:04:05")),
				verified,
			}
			startY := pdf.GetY()
			maxHeight := 0.0

			for j, txt := range row {
				lines := pdf.SplitLines([]byte(txt), colWidths[j])
				h := float64(len(lines)) * 5
				if h > maxHeight {
					maxHeight = h
				}
			}
			xStart := pdf.GetX()
			for j, txt := range row {
				x := pdf.GetX()
				y := pdf.GetY()

				pdf.MultiCell(colWidths[j], 5, txt, "1", "", false)
				pdf.SetXY(x+colWidths[j], y)
			}
			pdf.SetXY(xStart, startY+maxHeight)
		}

		var buf bytes.Buffer
		if err := pdf.Output(&buf); err != nil {
			return "", nil, "", response.ErrorBuilder(http.StatusInternalServerError, err, "server_error")
		}

		filename := "ISS CleanCare - Laporan Data Petugas Kebersihan.pdf"
		return filename, &buf, "pdf", nil

	} else {
		f := excelize.NewFile()
		sheet := "ISS CleanCare - Laporan Data Petugas Kebersihan"
		index, err := f.NewSheet(general.TruncateSheetName(sheet))
		if err != nil {
			return "", nil, "", response.ErrorBuilder(http.StatusInternalServerError, err, "server_error")
		}
		f.DeleteSheet("Sheet1")
		f.SetActiveSheet(index)
		f.SetCellValue(sheet, "A1", "No")
		f.SetCellValue(sheet, "B1", "Nomor ID")
		f.SetCellValue(sheet, "C1", "Nama")
		f.SetCellValue(sheet, "D1", "Email")
		f.SetCellValue(sheet, "E1", "Jabatan")
		f.SetCellValue(sheet, "F1", "Tanggal Terdaftar")
		f.SetCellValue(sheet, "G1", "Status Verifikasi")
		for i, v := range data {
			colA := fmt.Sprintf("A%d", i+2)
			colB := fmt.Sprintf("B%d", i+2)
			colC := fmt.Sprintf("C%d", i+2)
			colD := fmt.Sprintf("D%d", i+2)
			colE := fmt.Sprintf("E%d", i+2)
			colF := fmt.Sprintf("F%d", i+2)
			colG := fmt.Sprintf("G%d", i+2)
			no := i + 1
			f.SetCellValue(sheet, colA, no)
			f.SetCellValue(sheet, colB, v.NumberId)
			f.SetCellValue(sheet, colC, v.Name)
			if v.Email == nil {
				f.SetCellValue(sheet, colD, "-")
				f.SetCellValue(sheet, colG, "Belum")
			} else {
				f.SetCellValue(sheet, colD, *v.Email)
				f.SetCellValue(sheet, colG, "Sudah")
			}
			if v.RoleId == constant.ROLE_ID_STAFF {
				f.SetCellValue(sheet, colE, "Petugas Kebersihan")
			} else {
				f.SetCellValue(sheet, colE, "Supervisor")
			}
			f.SetCellValue(sheet, colF, general.ConvertDateTimeToIndonesian(v.CreatedAt.Format("2006-01-02 15:04:05")))
		}

		var buf bytes.Buffer
		if err := f.Write(&buf); err != nil {
			return "", nil, "", response.ErrorBuilder(http.StatusInternalServerError, err, "server_error")
		}
		filename := "ISS CleanCare - Laporan Data Petugas Kebersihan.xlsx"
		return filename, &buf, "excel", nil
	}
}
