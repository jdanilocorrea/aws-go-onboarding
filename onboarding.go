package main

import (
	"bufio"
	"encoding/base64"
	"fmt"
	"log"
	"math/rand"
	"os"
	"path/filepath"
	"regexp"
	"strconv"
	"strings"
	"time"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/iam"
)

// Função para ler o arquivo .aws/config e extrair os profiles, regiões e endpoint URLs correspondentes
func readAWSConfig() map[string]map[string]string {
	homeDir, err := os.UserHomeDir()
	if err != nil {
		fmt.Println("Erro ao obter diretório home:", err)
		os.Exit(1)
	}

	configFile := filepath.Join(homeDir, ".aws", "config")
	file, err := os.Open(configFile)
	if err != nil {
		fmt.Println("Erro ao abrir o arquivo .aws/config:", err)
		os.Exit(1)
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
		fmt.Println("Erro ao ler o arquivo .aws/config:", err)
		os.Exit(1)
	}

	return config
}

// Função para configurar a sessão AWS com base no profile escolhido e no endpoint URL
func configureAWS(profile string, config map[string]string) *session.Session {
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
		fmt.Println("Erro ao configurar a sessão da AWS:", err)
		os.Exit(1)
	}
	return sess
}

// Função para criar um usuário IAM
func createUser(sess *session.Session) string {
	svc := iam.New(sess)

	reader := bufio.NewReader(os.Stdin)
	fmt.Print("Digite o nome do usuário que deseja criar: ")
	userName, _ := reader.ReadString('\n')
	userName = strings.TrimSpace(userName)

	_, err := svc.CreateUser(&iam.CreateUserInput{
		UserName: aws.String(userName),
	})
	if err != nil {
		log.Fatal("Erro ao criar usuário:", err)
	}
	fmt.Println("Usuário criado com sucesso:", userName)
	return userName
}

// Função para listar grupos IAM existentes e retornar o grupo selecionado
func listGroups(sess *session.Session) *iam.Group {
	svc := iam.New(sess)

	result, err := svc.ListGroups(nil)
	if err != nil {
		log.Fatal("Erro ao listar grupos:", err)
	}

	groups := result.Groups
	fmt.Println("Grupos existentes:")
	for i, group := range groups {
		fmt.Printf("%d. %s\n", i+1, *group.GroupName)
	}

	reader := bufio.NewReader(os.Stdin)
	fmt.Print("Escolha o número do grupo a ser adicionado ao usuário: ")
	choiceStr, _ := reader.ReadString('\n')
	choiceStr = strings.TrimSpace(choiceStr)
	choice, err := strconv.Atoi(choiceStr)
	if err != nil || choice < 1 || choice > len(groups) {
		fmt.Println("Escolha inválida!")
		os.Exit(1)
	}

	return groups[choice-1]
}

// Função para adicionar usuário ao grupo selecionado
func addUserToGroup(sess *session.Session, userName string, group *iam.Group) {
	svc := iam.New(sess)

	_, err := svc.AddUserToGroup(&iam.AddUserToGroupInput{
		GroupName: aws.String(*group.GroupName),
		UserName:  aws.String(userName),
	})
	if err != nil {
		log.Fatal("Erro ao adicionar usuário ao grupo:", err)
	}
	fmt.Printf("Usuário %s adicionado ao grupo %s com sucesso.\n", userName, *group.GroupName)
}

// Função para anexar uma política ao usuário pelo nome da política
func attachPolicyByName(sess *session.Session, userName string, policyName string) {
	svc := iam.New(sess)

	// Obtém o ARN da política pelo nome
	arn, err := getPolicyARNByName(svc, policyName)
	if err != nil {
		log.Fatalf("Erro ao obter ARN da política: %v", err)
	}

	// Anexa a política ao usuário
	_, err = svc.AttachUserPolicy(&iam.AttachUserPolicyInput{
		UserName:  aws.String(userName),
		PolicyArn: aws.String(arn),
	})
	if err != nil {
		log.Fatalf("Erro ao anexar política ao usuário: %v", err)
	}

	fmt.Println("Política adicionada com sucesso ao usuário:", userName)
}

// Função para obter o ARN da política pelo nome
func getPolicyARNByName(svc *iam.IAM, policyName string) (string, error) {
	input := &iam.ListPoliciesInput{}

	var policyArn string

	err := svc.ListPoliciesPages(input, func(page *iam.ListPoliciesOutput, lastPage bool) bool {
		for _, policy := range page.Policies {
			if *policy.PolicyName == policyName {
				policyArn = *policy.Arn
				return false // Interrompe a iteração assim que a política é encontrada
			}
		}
		return true // Continua a iteração para a próxima página
	})

	if err != nil {
		return "", err
	}

	if policyArn == "" {
		return "", fmt.Errorf("policy not found: %s", policyName)
	}

	return policyArn, nil
}

// createConsoleAccess cria acesso ao console AWS para o novo usuário
func createConsoleAccess(sess *session.Session, userName string) {
	// Crie um cliente IAM a partir da sessão
	svc := iam.New(sess)

	// Gere uma senha aleatória temporária
	tempPassword, err := generateRandomPassword()
	if err != nil {
		log.Fatalf("Erro ao gerar a senha temporária: %v", err)
	}

	// Crie um perfil de login para o usuário com a senha temporária
	_, err = svc.CreateLoginProfile(&iam.CreateLoginProfileInput{
		UserName:              aws.String(userName),
		Password:              aws.String(tempPassword),
		PasswordResetRequired: aws.Bool(true),
	})
	if err != nil {
		log.Fatalf("Erro ao criar acesso ao console da AWS: %v", err)
	}

	// // Verifique se o resultado e o perfil de login não são nulos
	// if result != nil && result.LoginProfile != nil {
	// 	fmt.Println("Perfil de login criado com sucesso:")
	// 	fmt.Println("User name:", *result.LoginProfile.UserName)
	// 	fmt.Println("Password reset required:", *result.LoginProfile.PasswordResetRequired)
	// } else {
	// 	fmt.Println("Erro: O objeto result ou result.LoginProfile é nil.")
	// }

	// Crie o conteúdo do arquivo console-access
	consoleAccessInfo := fmt.Sprintf("Link: [%s]\nUser: %s\nPassword: %s\n", "http://profile", userName, tempPassword)

	// Abrir o arquivo no modo de anexo para escrita
	file, err := os.OpenFile("console-access", os.O_APPEND|os.O_WRONLY|os.O_CREATE, 0644)
	if err != nil {
		log.Fatalf("Erro ao abrir o arquivo 'console-access': %v", err)
	}
	defer file.Close()

	// Escrever no arquivo
	_, err = file.WriteString(consoleAccessInfo)
	if err != nil {
		log.Fatalf("Erro ao escrever no arquivo 'console-access': %v", err)
	}

	fmt.Println("Acesso ao console da AWS criado com sucesso para o usuário:", userName)
	fmt.Println("Arquivo 'console-access' criado com sucesso.")
}

// generateRandomPassword gera uma senha aleatória segura
func generateRandomPassword() (string, error) {
	// Defina o tamanho da senha
	const passwordLength = 12

	// Crie um slice de bytes para armazenar a senha
	password := make([]byte, passwordLength)

	// Use o tempo atual como semente para a função de geração de números aleatórios
	rand.Seed(time.Now().UnixNano())

	// Gere números aleatórios para cada caractere da senha
	for i := 0; i < passwordLength; i++ {
		// Gere um número aleatório entre 33 e 126, que corresponde aos caracteres imprimíveis ASCII
		password[i] = byte(rand.Intn(94) + 33)
	}

	// Converta os bytes em uma string base64
	return base64.StdEncoding.EncodeToString(password), nil
}

func main() {
	fmt.Println("Opções disponíveis baseadas no arquivo .aws/config:")
	awsConfig := readAWSConfig()
	i := 1
	for profile, config := range awsConfig {
		fmt.Printf("%d. Profile: %s, Região: %s, Endpoint URL: %s\n", i, profile, config["region"], config["endpoint_url"])
		i++
	}

	reader := bufio.NewReader(os.Stdin)
	fmt.Print("Escolha o número do profile AWS desejado: ")
	choiceStr, _ := reader.ReadString('\n')
	choiceStr = strings.TrimSpace(choiceStr)
	choice, err := strconv.Atoi(choiceStr)
	if err != nil || choice < 1 || choice > len(awsConfig) {
		fmt.Println("Escolha inválida!")
		os.Exit(1)
	}

	var selectedProfile string
	for profile := range awsConfig {
		if choice == 1 {
			selectedProfile = profile
			break
		}
		choice--
	}

	// Configurar sessão AWS com base no profile escolhido e no endpoint URL
	sess := configureAWS(selectedProfile, awsConfig[selectedProfile])

	// Exibir informações da sessão
	fmt.Println("Informações da sessão:")
	fmt.Println("Profile:", selectedProfile)
	fmt.Println("Região:", *sess.Config.Region)
	fmt.Println("Endpoint URL:", *sess.Config.Endpoint)

	// Criar usuário
	userName := createUser(sess)

	// Anexar política ao usuário pelo nome da política
	attachPolicyByName(sess, userName, "SelfManageMFADevice")

	// Listar grupos existentes
	group := listGroups(sess)

	// Adicionar usuário ao grupo selecionado
	addUserToGroup(sess, userName, group)

	// Criar acesso ao console AWS para o novo usuário
	createConsoleAccess(sess, userName)
}
