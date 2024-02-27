# Use a imagem oficial do Golang como base
FROM golang:latest

# Configure o diretório de trabalho dentro do container
WORKDIR /go/src/app

# Copie os arquivos necessários para o diretório de trabalho
COPY . .

# Baixe as dependências do Go
RUN go mod download

# Compile o código Go
RUN go build -o main .

# Exponha a porta necessária para o aplicativo
EXPOSE 8080

# Comando padrão para executar o aplicativo
CMD ["./main"]
