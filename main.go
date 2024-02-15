package main

import (
	"bufio"
	"fmt"
	"log"
	"os"
	"strconv"
	"strings"

	. "github.com/jdanilocorrea/aws-go-onboarding/add_groups_list"
	. "github.com/jdanilocorrea/aws-go-onboarding/add_policy"
	. "github.com/jdanilocorrea/aws-go-onboarding/aws_cli_access"
	. "github.com/jdanilocorrea/aws-go-onboarding/console_access"
	. "github.com/jdanilocorrea/aws-go-onboarding/create_user"
	. "github.com/jdanilocorrea/aws-go-onboarding/sess_config"
)

func main() {

	// Ler as configurações AWS
	awsConfig := ReadAWSConfig()

	// Ordenar os perfis
	profiles := SortProfiles(awsConfig)

	// Exibir opções disponíveis baseadas nos perfis AWS
	fmt.Println("Opções disponíveis baseadas no arquivo .aws/config:")
	for i, profile := range profiles {
		config := awsConfig[profile]
		fmt.Printf("[%d]. Profile: %s, Região: %s, Endpoint URL: %s\n", i+1, profile, config["region"], config["endpoint_url"])
	}

	reader := bufio.NewReader(os.Stdin)
	fmt.Print("Escolha o número do profile AWS desejado: ")
	choiceStr, _ := reader.ReadString('\n')
	choiceStr = strings.TrimSpace(choiceStr)
	choice, err := strconv.Atoi(choiceStr)
	if err != nil || choice < 1 || choice > len(profiles) {
		fmt.Println("Escolha inválida!")
		os.Exit(1)
	}

	selectedProfile := profiles[choice-1]

	// Configurar sessão AWS com base no profile escolhido
	sess := ConfigureAWS(selectedProfile, awsConfig[selectedProfile])

	// Exibir informações da sessão
	fmt.Println("Informações da sessão:")
	fmt.Println("Profile:", selectedProfile)
	fmt.Println("Região:", *sess.Config.Region)
	fmt.Println("Endpoint URL:", *sess.Config.Endpoint)

	// Criar usuário
	userName := CreateUser(sess)

	// Anexar política ao usuário pelo nome da política
	AttachPolicyByName(sess, userName, "SelfManageMFADevice")

	// Listar grupos existentes
	groups := ListGroups(sess)

	// Adicionar usuário ao grupo selecionado
	AddUserToGroups(sess, userName, groups)

	// Criar acesso ao console AWS para o novo usuário
	CreateConsoleAccess(sess, selectedProfile, userName)

	// Criar acesso ao AWS CLI para o novo usuário
	accessKeyID, secretAccessKey, err := CreateAWSCLIUserAccess(sess, userName)
	if err != nil {
		log.Fatalf("Erro ao criar acesso ao AWS CLI para o usuário %s: %v", userName, err)
	}

	// Gerar o arquivo "credentials" com as chaves obtidas
	GenerateCredentialsFile(selectedProfile, accessKeyID, secretAccessKey)

	fmt.Println("Processo concluído com sucesso!")
}
