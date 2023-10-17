# Distrlocker

The README is also available in [english](README.md).

---

A biblioteca Distrlocker, escrita em Golang, proporciona uma maneira simples e eficaz de gerenciar locks distribuídos
por meio do Redis. Locks distribuídos são uma primitiva valiosa em ambientes nos quais diversos processos precisam
operar com recursos compartilhados de maneira exclusiva.

## Instalação

Para facilitar a integração com o Redis, esta biblioteca utiliza como dependência a biblioteca oficial do serviço
([github.com/redis/go-redis](https://github.com/redis/go-redis)). Para incorporar os pacotes em seu projeto, execute os seguintes comandos:

```bash
go get -u github.com/viniciuscrisol/distrlocker
go get -u github.com/redis/go-redis/v9
```

## Exemplo

O seguinte exemplo ilustra um caso de uso simples para aquisição e liberação de um lock:

```go
func main() {
    // Criação de um cliente Redis
    clienteRedis := redis.NewClient(
        &redis.Options{Addr: "127.0.0.1:6379", WriteTimeout: time.Second * 3},
    )

    // Criação do locker distribuído com o timeout de 5000 ms
    locker := distrlocker.NewDistrLocker(5000, clienteRedis)

    // Obtenção do lock
    meuLock, err := locker.Acquire("minha-chave")
    if err != nil {
        fmt.Println("Falha ao adquirir o lock:", err)
        return
    }
    // Liberação do lock. Caso os 5000 ms estipulados
    // expirem, o lock será liberado automaticamente
    defer meuLock.Release()

    // Processamento intensivo...
}
```

## Testes

A biblioteca inclui um conjunto de testes que simulam cenários de aquisição e liberação de locks. Para executar os
testes, é necessário ter o Docker instalado. Além disso, deve-se garantir que a porta 6379 esteja disponível.
Ao executar o seguinte comando, a cobertura será obtida:

```bash
go test ./... -cover -coverprofile=coverage
```
