# Zaap

Zaap é um projeto open-source desenvolvido em Go que fornece uma API moderna e leve para gerenciar sessões de WhatsApp, enviar e receber mensagens, processar eventos em tempo real e automatizar interações. O Zaap foi criado para ser simples de usar, escalável e totalmente extensível, permitindo que desenvolvedores construam bots, integrações e sistemas de comunicação confiáveis com múltiplas sessões simultâneas.

Inspirado em projetos como o (EvolutionAPI)[https://github.com/EvolutionAPI/evolution-api], o Zaap utiliza a biblioteca (whatsmeow)[go.mau.fi/whatsmeow] para se conectar ao WhatsApp.

## Features

- **Gerenciamento de Múltiplas Sessões**: Crie e gerencie múltiplas sessões de WhatsApp de forma independente.
- **API RESTful**: Interface HTTP para criar sessões, obter QR Code, enviar mensagens e mais.
- **Tempo Real com WebSockets**: Receba eventos do WhatsApp (mensagens, status, etc.) em tempo real.
- **Armazenamento Persistente**: As sessões são salvas no disco e restauradas automaticamente.
- **Configuração Simples**: Um único arquivo `config.yml` para configurar o servidor.
- **Extensível**: O código é modular e fácil de estender.

## Começando

### Pré-requisitos

- Go 1.18 ou superior instalado.

### Instalação

1.  Clone o repositório:
    ```bash
    git clone https://github.com/gedsonn/zaapi.git
    cd zaapi
    ```

2.  Instale as dependências:
    ```bash
    go mod tidy
    ```

3.  Compile e execute o projeto:
    ```bash
    go run main.go
    ```

O servidor será iniciado na porta especificada no `config.yml` (padrão: `8080`).

## Endpoints da API

A seguir estão os principais endpoints da API.

### Sessões

-   `POST /sessions`: Cria uma nova sessão.
    -   **Resposta**:
        ```json
        {
          "message": "Sessão criada com sucesso",
          "session_id": "1994603210114863104"
        }
        ```

-   `GET /sessions/:session/qr`: Obtém o QR Code para parear um dispositivo.
    -   **Parâmetros de URL**:
        -   `session`: ID da sessão.
    -   **Resposta**:
        ```json
        {
          "base64": "data:image/png;base64,iVBORw0KGgo...",
          "expires_in": 1199
        }
        ```

## Configuração

O servidor é configurado através do arquivo `config.yml`. Se o arquivo não existir, um será criado com os valores padrão na primeira vez que o aplicativo for executado.

```yaml
server:
  enable: true
  host: 0.0.0.0
  port: 8080
```

-   `server.enable`: `true` para habilitar o servidor HTTP.
-   `server.host`: O host no qual o servidor irá escutar.
-   `server.port`: A porta na qual o servidor irá escutar.

## Contribuição

Contribuições são bem-vindas! Sinta-se à vontade para abrir uma issue ou enviar um pull request.

## Licença

Este projeto é licenciado sob a Licença MIT. Veja o arquivo `LICENSE` para mais detalhes.
