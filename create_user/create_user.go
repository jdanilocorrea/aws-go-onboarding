package create_user

import (
	"bufio"
	"fmt"
	"log"
	"os"
	"strings"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/iam"
)

// Função para criar um usuário IAM
func CreateUser(sess *session.Session) string {
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
