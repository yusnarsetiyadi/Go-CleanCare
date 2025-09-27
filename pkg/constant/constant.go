package constant

const (
	DRIVE_FOLDER = "ISS_CleanCare_App"

	ROLE_ID_ADMIN = 1
	ROLE_ID_STAFF = 2

	REDIS_REQUEST_IP_KEYS      = "iss-reset-password:ip:%s"
	REDIS_REQUEST_MAX_ATTEMPTS = 5
	REDIS_REQUEST_IP_EXPIRE    = 240
	REDIS_KEY_USER_LOGIN       = "iss_login_token_user_"
	REDIS_KEY_AUTO_LOGOUT      = "iss_user_auto_logout"
	REDIS_KEY_REFRESH_TOKEN    = "iss-refresh-token:%s"
	REDIS_MAX_REFRESH_TOKEN    = 30

	PATH_FILE_SAVED    = "../file_saved"
	PATH_ASSETS_IMAGES = "assets/images"
	PATH_SHARE         = "/var/www/html/iss_cleancare/share"
)

var (
	BASE_URL          string = ""
	BASE_URL_FRONTEND string = ""
)
