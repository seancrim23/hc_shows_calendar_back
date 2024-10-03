package utils

import "errors"

const SHOW_COLLECTION = "show"
const USER_COLLECTION = "user"
const VERIFICATION_COLLECTION = "verification"

const APPLICATION_PORT = "APPLICATION_PORT"
const SITE_KEY = "SITE_KEY"
const ALLOWED_ORIGINS = "ALLOWED_ORIGINS"
const FIRESTORE_EMULATOR_HOST = "FIRESTORE_EMULATOR_HOST"
const GCP_PROJECT_ID = "GCP_PROJECT_ID"

const LOCALHOST = "localhost"
const FIRESTORE_EMULATOR_PORT = "5050"

const FROM = "From"
const TO = "To"
const SUBJECT = "Subject"
const TEXT_HTML = "text/html"
const GMAIL_SMTP = "smtp.gmail.com"
const GMAIL_SMTP_PORT = 587

// TODO maybe replace at some point with a proper domain handled email address instead of free gmail
const MY_EMAIL = "theseancrim@gmail.com"
const GMAIL_AUTH = "GMAIL_AUTH"

var ErrUnauthorized = errors.New("failed authorization")
var ErrTokenGeneration = errors.New("error generating token")
var ErrUserDoesntExist = errors.New("cannot find user")
var ErrShowDoesntExist = errors.New("cannot find show")

// TODO take another look at this, maybe more descriptive?
var ErrUserDataMalformed = errors.New("user data is malformed")
var ErrShowDataMalformed = errors.New("show data is malformed")

var ADMIN_USER = "admin"
var SECRET_USER_KEY = "hc_show_calendar_user_key"

var GOOGLE_URL = "GOOGLE_URL"
var TEMPLATE_BUCKET = "TEMPLATE_BUCKET"
