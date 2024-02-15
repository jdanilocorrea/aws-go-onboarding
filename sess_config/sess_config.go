package sess_config

import (
	"bufio"
	"log"
	"os"
	"path/filepath"
	"regexp"
	"sort"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/session"
)

const (
	awsConfigFile = ".aws/config"
)

// Função para ler o arquivo .aws/config e extrair os profiles, regiões e endpoint URLs correspondentes
func ReadAWSConfig() map[string]map[string]string {
	homeDir, err := os.UserHomeDir()
	if err != nil {
		log.Fatalf("Erro ao obter diretório home: %v", err)
	}

	configFile := filepath.Join(homeDir, awsConfigFile)
	file, err := os.Open(configFile)
	if err != nil {
		log.Fatalf("Erro ao abrir o arquivo %s: %v", awsConfigFile, err)
	}
	defer file.Close()

	config := make(map[string]map[string]string)
	profileRegex := regexp.MustCompile(`^\s*\[([^\]]+)\]\s*$`)
	regionRegex := regexp.MustCompile(`^\s*region\s*=\s*([^\s]+)\s*$`)
	endpointURLRegex := regexp.MustCompile(`^\s*endpoint_url\s*=\s*([^\s]+)\s*$`)

	var currentProfile string
	var currentConfig map[string]string

	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		line := scanner.Text()

		if matches := profileRegex.FindStringSubmatch(line); len(matches) > 0 {
			currentProfile = matches[1]
			currentConfig = make(map[string]string)
			config[currentProfile] = currentConfig
		} else if matches := regionRegex.FindStringSubmatch(line); len(matches) > 0 {
			currentConfig["region"] = matches[1]
		} else if matches := endpointURLRegex.FindStringSubmatch(line); len(matches) > 0 {
			currentConfig["endpoint_url"] = matches[1]
		}
	}

	if err := scanner.Err(); err != nil {
		log.Fatalf("Erro ao ler o arquivo %s: %v", awsConfigFile, err)
	}

	return config
}

// Função para configurar a sessão AWS com base no profile escolhido e no endpoint URL
func ConfigureAWS(profile string, config map[string]string) *session.Session {
	// Cria uma nova sessão usando as configurações compartilhadas do arquivo .aws/config
	endpointURL := config["endpoint_url"]
	sess, err := session.NewSessionWithOptions(session.Options{
		Profile:           profile,
		SharedConfigState: session.SharedConfigEnable,
		Config: aws.Config{
			Endpoint: &endpointURL,
		},
	})
	if err != nil {
		log.Fatalf("Erro ao configurar a sessão da AWS para o profile %s: %v", profile, err)
	}
	return sess
}

// Função para ordenar os perfis
func SortProfiles(profiles map[string]map[string]string) []string {
	keys := make([]string, 0, len(profiles))
	for key := range profiles {
		keys = append(keys, key)
	}
	sort.Strings(keys)
	return keys
}
