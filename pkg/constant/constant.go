package constant

const (
	DRIVE_FOLDER = "CleanCare_App"

	ROLE_ID_ADMIN                             = 1
	ROLE_ID_STAFF                             = 2
	TASK_ID_DAILY                             = 1
	TASK_ID_SERVICE                           = 2
	REDIS_REQUEST_RESET_PASSWORD_IP_KEYS      = "cleancare-reset-password:ip:%s"
	REDIS_REQUEST_VERIFY_NUMBER_IP_KEYS       = "cleancare-verify-mumber:ip:%s"
	REDIS_REQUEST_MAX_ATTEMPTS_RESET_PASSWORD = 5
	REDIS_REQUEST_MAX_ATTEMPTS_VERIFY_NUMBER  = 10
	REDIS_REQUEST_IP_EXPIRE                   = 240
	REDIS_KEY_USER_LOGIN                      = "cleancare_login_token_user_"
	REDIS_KEY_AUTO_LOGOUT                     = "cleancare_user_auto_logout"
	REDIS_KEY_REFRESH_TOKEN                   = "cleancare-refresh-token:%s"
	REDIS_KEY_UNREAD_COMMENT                  = "cleancare-unread-comment:%d"
	REDIS_MAX_REFRESH_TOKEN                   = 30

	PATH_FILE_SAVED    = "../file_saved"
	PATH_ASSETS_IMAGES = "assets/images"
	PATH_SHARE         = "/var/www/html/cleancare/share"
)

var (
	BASE_URL          string = ""
	BASE_URL_FRONTEND string = ""
)
