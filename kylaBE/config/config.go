package config

import (
	"log"
	"os"
	"path/filepath"
	"strconv"

	"github.com/spf13/viper"
)

type PostgresConfig struct {
	PostgresHost string `mapstructure:"POSTGRES_HOST"`
	PostgresPort string `mapstructure:"POSTGRES_PORT"`
	PostgresDB   string `mapstructure:"POSTGRES_DB"`
	PostgresUser string `mapstructure:"POSTGRES_USER"`
	PostgresPass string `mapstructure:"POSTGRES_PASS"`
}

type EnvConfigs struct {
	Environment          string `mapstructure:"ENVIRONMENT"`
	Port                 string `mapstructure:"PORT"`
	JwtSecret            string `mapstructure:"JWT_SECRET_KEY"`
	AuthSvcUrl           string `mapstructure:"AUTH_SVC_URL"`
	NatsURL              string `mapstructure:"NATS_URL"`
	WhatsAppWebhookSecret string `mapstructure:"WA_WEBHOOK_SECRET"`
	WhatsAppVerifyToken  string `mapstructure:"WA_VERIFY_TOKEN"`
	WhatsAppAccessToken   string `mapstructure:"WA_ACCESS_TOKEN"`
	WhatsAppPhoneNumberID string `mapstructure:"WA_PHONE_NUMBER_ID"`

	IMAPHost             string `mapstructure:"IMAP_HOST"`
	IMAPPort             string `mapstructure:"IMAP_PORT"`
	IMAPUser             string `mapstructure:"IMAP_USER"`
	IMAPPassword         string `mapstructure:"IMAP_PASS"`
	IMAPPollIntervalSecs int    `mapstructure:"IMAP_POLL_INTERVAL_SECS"`

	SMTPHost     string `mapstructure:"SMTP_HOST"`
	SMTPPort     string `mapstructure:"SMTP_PORT"`
	SMTPUser     string `mapstructure:"SMTP_USER"`
	SMTPPassword string `mapstructure:"SMTP_PASS"`

	SMSProvider string `mapstructure:"SMS_PROVIDER"`

	TwilioAccountSID string `mapstructure:"TWILIO_ACCOUNT_SID"`
	TwilioAuthToken  string `mapstructure:"TWILIO_AUTH_TOKEN"`
	TwilioFrom       string `mapstructure:"TWILIO_FROM"`

	AfricasTalkingAPIKey   string `mapstructure:"AT_API_KEY"`
	AfricasTalkingUsername string `mapstructure:"AT_USERNAME"`
	AfricasTalkingFrom     string `mapstructure:"AT_FROM"`

	SLAScanIntervalSecs int `mapstructure:"SLA_SCAN_INTERVAL_SECS"`
}

type RsConfig struct {
	//apiKey string, fromEmail string, supportEmail string, baseURL string
	ResendApiKey       string `mapstructure:"RESEND_API_KEY"`
	ResendFromEmail    string `mapstructure:"RESEND_FROM_EMAIL"`
	ResendSupportEmail string `mapstructure:"RESEND_SUPPORT_EMAIL"`
	ResendBaseURL      string `mapstructure:"RESEND_BASE_URL"`
}

type FirebaseConfig struct {
	FbCredentials string `mapstructure:"FB_CREDENTIALS"`
}

type WebAuthnConfig struct {
	RPID          string `mapstructure:"WEB_AUTHN_RP_ID"`
	RPOrigin      string `mapstructure:"WEB_AUTHN_RP_ORIGIN"`
	RPDisplayName string `mapstructure:"WEB_AUTHN_RP_DISPLAY_NAME"`
}

type RedisConfig struct {
	Addr     string `mapstructure:"REDIS_ADDR"`
	Password string `mapstructure:"REDIS_PASSWORD"`
	DB       int    `mapstructure:"REDIS_DB"`
}

type AwsCredentialsConfig struct {
	AwsRegion    string `mapstructure:"AWS_REGION"`
	AwsAccessKey string `mapstructure:"AWS_ACCESS_KEY"`
	AwsSecretKey string `mapstructure:"AWS_SECRET_KEY"`
}

type Config struct {
	EnvConfigs           EnvConfigs
	RsConfig             RsConfig
	PostgresConfig       PostgresConfig
	Clients              Clients
	FirebaseConfig       FirebaseConfig
	WebAuthnConfig       WebAuthnConfig
	RedisConfig          RedisConfig
	AwsCredentialsConfig AwsCredentialsConfig
}

type Clients struct {
	LeadsAddress string `mapstructure:"LEADS_ADDRESS"`
}

func isSet(s string) bool {
	return len(s) > 0
}

func LoadConfig() (*Config, error) {
	config := &Config{}
	port, jwtSecretKey, environment := os.Getenv("PORT"), os.Getenv("JWT_SECRET_KEY"), os.Getenv("ENVIRONMENT")
	postgresHost, postgresPort, postgresDB, postgresUser, postgresPass := os.Getenv("POSTGRES_HOST"), os.Getenv("POSTGRES_PORT"), os.Getenv("POSTGRES_DB"), os.Getenv("POSTGRES_USER"), os.Getenv("POSTGRES_PASS")
	resendapikey, resendBaseUrl, resendFromEmail, resendSupportEmail := os.Getenv("RESEND_API_KEY"), os.Getenv("RESEND_BASE_URL"), os.Getenv("RESEND_FROM_EMAIL"), os.Getenv("RESEND_SUPPORT_EMAIL")
	leadsAddress := os.Getenv("LEADS_ADDRESS")
	webAuthnRPID, webAuthnRPOrigin, webAuthnRPDisplayName := os.Getenv("WEB_AUTHN_RP_ID"), os.Getenv("WEB_AUTHN_RP_ORIGIN"), os.Getenv("WEB_AUTHN_RP_DISPLAY_NAME")
	awsRegion, awsAccessKey, awsSecretKey := os.Getenv("AWS_REGION"), os.Getenv("AWS_ACCESS_KEY"), os.Getenv("AWS_SECRET_KEY")

	fbCredentials := os.Getenv("FB_CREDENTIALS")
	waWebhookSecret := os.Getenv("WA_WEBHOOK_SECRET")
	waVerifyToken := os.Getenv("WA_VERIFY_TOKEN")

	if !isSet(port) && !isSet(jwtSecretKey) && !isSet(postgresHost) && !isSet(postgresPort) && !isSet(postgresDB) && !isSet(postgresUser) && !isSet(postgresPass) && !isSet(environment) && !isSet(resendapikey) && !isSet(resendBaseUrl) && !isSet(resendFromEmail) && !isSet(leadsAddress) && !isSet(fbCredentials) && !isSet(awsRegion) && !isSet(awsAccessKey) && !isSet(awsSecretKey) {
		cwd, err := os.Getwd()
		if err != nil {
			log.Fatal("Error getting current working directory:", err)
		}
		configPath := filepath.Join(cwd, "envs")

		viper.AddConfigPath(configPath)
		viper.SetConfigName(".env")
		viper.SetConfigType("env")

		viper.AutomaticEnv()
		// Read the config file
		if err := viper.ReadInConfig(); err != nil {
			log.Fatalf("Error reading config file, %s", err)
		}
		if err = viper.Unmarshal(&config.EnvConfigs); err != nil {
			log.Fatalf("Unable to decode into struct, %v", err)
		}
		if err = viper.Unmarshal(&config.RsConfig); err != nil {
			log.Fatalf("Unable to decode into struct, %v", err)
		}
		if err = viper.Unmarshal(&config.PostgresConfig); err != nil {
			log.Fatalf("Unable to decode into struct, %v", err)
		}
		if err = viper.Unmarshal(&config.Clients); err != nil {
			log.Fatalf("Unable to decode into struct, %v", err)
		}
		if err = viper.Unmarshal(&config.FirebaseConfig); err != nil {
			log.Fatalf("Unable to decode into struct, %v", err)
		}
		if err = viper.Unmarshal(&config.WebAuthnConfig); err != nil {
			log.Fatalf("Unable to decode into struct, %v", err)
		}

		return config, nil
	} else {
		log.Println("Port2 is: ", port)
		config = &Config{
			EnvConfigs: EnvConfigs{
				Environment:           environment,
				Port:                  port,
				JwtSecret:             jwtSecretKey,
				WhatsAppWebhookSecret: waWebhookSecret,
				WhatsAppVerifyToken:   waVerifyToken,
				WhatsAppAccessToken:   os.Getenv("WA_ACCESS_TOKEN"),
				WhatsAppPhoneNumberID: os.Getenv("WA_PHONE_NUMBER_ID"),

				IMAPHost:             os.Getenv("IMAP_HOST"),
				IMAPPort:             os.Getenv("IMAP_PORT"),
				IMAPUser:             os.Getenv("IMAP_USER"),
				IMAPPassword:         os.Getenv("IMAP_PASS"),
				IMAPPollIntervalSecs: getIntEnv("IMAP_POLL_INTERVAL_SECS", 60),

				SMTPHost:     os.Getenv("SMTP_HOST"),
				SMTPPort:     os.Getenv("SMTP_PORT"),
				SMTPUser:     os.Getenv("SMTP_USER"),
				SMTPPassword: os.Getenv("SMTP_PASS"),

				SMSProvider: os.Getenv("SMS_PROVIDER"),

				TwilioAccountSID: os.Getenv("TWILIO_ACCOUNT_SID"),
				TwilioAuthToken:  os.Getenv("TWILIO_AUTH_TOKEN"),
				TwilioFrom:       os.Getenv("TWILIO_FROM"),

				AfricasTalkingAPIKey:   os.Getenv("AT_API_KEY"),
				AfricasTalkingUsername: os.Getenv("AT_USERNAME"),
				AfricasTalkingFrom:     os.Getenv("AT_FROM"),

				SLAScanIntervalSecs: getIntEnv("SLA_SCAN_INTERVAL_SECS", 60),
			},
			RsConfig: RsConfig{
				ResendApiKey:       resendapikey,
				ResendBaseURL:      resendBaseUrl,
				ResendFromEmail:    resendFromEmail,
				ResendSupportEmail: resendSupportEmail,
			},
			PostgresConfig: PostgresConfig{
				PostgresHost: postgresHost,
				PostgresPort: postgresPort,
				PostgresDB:   postgresDB,
				PostgresUser: postgresUser,
				PostgresPass: postgresPass,
			},
			Clients: Clients{
				LeadsAddress: leadsAddress,
			},
			FirebaseConfig: FirebaseConfig{
				FbCredentials: fbCredentials,
			},
			WebAuthnConfig: WebAuthnConfig{
				RPID:          webAuthnRPID,
				RPOrigin:      webAuthnRPOrigin,
				RPDisplayName: webAuthnRPDisplayName,
			},
			AwsCredentialsConfig: AwsCredentialsConfig{
				AwsRegion:    awsRegion,
				AwsAccessKey: awsAccessKey,
				AwsSecretKey: awsSecretKey,
			},
		}
		return config, nil
	}
}

func getIntEnv(key string, defaultVal int) int {
	val := os.Getenv(key)
	if val == "" {
		return defaultVal
	}
	i, err := strconv.Atoi(val)
	if err != nil {
		return defaultVal
	}
	return i
}
