#

NAME = Counter
PACKAGE = github.com/ChizhovVadim/CounterGo

all:
	GOOS=linux   GOARCH=amd64 go build -o $(NAME)-linux-64       $(PACKAGE)
	GOOS=windows GOARCH=amd64 go build -o $(NAME)-windows-64.exe $(PACKAGE)

