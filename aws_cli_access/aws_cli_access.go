package aws_cli_access

import (
	"fmt"
	"log"
	"os"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/iam"
)

// Função para criar acesso ao AWS CLI para o novo usuário
func CreateAWSCLIUserAccess(sess *session.Session, userName string) (string, string, error) {
	fmt.Println("Criando acesso ao AWS CLI para o usuário:", userName)

	svc := iam.New(sess)

	// Criar o acesso ao AWS CLI para o usuário
	result, err := svc.CreateAccessKey(&iam.CreateAccessKeyInput{
		UserName: aws.String(userName),
	})
	if err != nil {
		return "", "", fmt.Errorf("Erro ao criar acesso ao AWS CLI para o usuário %s: %v", userName, err)
	}

	// Extrair as chaves de acesso
	accessKeyID := *result.AccessKey.AccessKeyId
	secretAccessKey := *result.AccessKey.SecretAccessKey

	fmt.Println("Acesso ao AWS CLI criado com sucesso para o usuário:", userName)

	return accessKeyID, secretAccessKey, nil
}

// Função para gerar o arquivo "credentials" com as chaves obtidas
func GenerateCredentialsFile(profile, accessKeyID, secretAccessKey string) {
	fmt.Println("Gerando arquivo 'credentials'...")

	credentialsContent := fmt.Sprintf("[%s]\naws_access_key_id = %s\naws_secret_access_key = %s\n", profile, accessKeyID, secretAccessKey)

	// Abrir o arquivo no modo de anexo para escrita
	file, err := os.OpenFile("credentials", os.O_APPEND|os.O_WRONLY|os.O_CREATE, 0644)
	if err != nil {
		log.Fatalf("Erro ao abrir o arquivo 'console-access': %v", err)
	}
	defer file.Close()

	// Escrever no arquivo
	_, err = file.WriteString(credentialsContent)
	if err != nil {
		log.Fatalf("Erro ao escrever no arquivo 'credentials': %v", err)
	}

	fmt.Println("Arquivo 'credentials' criado com sucesso.")
}
