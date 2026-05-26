# go.raphai.eu

Redirector de URLs leve e rápido, com painel administrativo e suporte a parâmetros UTM para rastreamento no Google Analytics.

## Funcionalidades

- **Redirect com loading** — tela com spinner animado enquanto o redirect acontece (meta refresh com delay 0, SEO-friendly)
- **Redirect 301 direto** — sem loading, redirect permanente e imediato
- **Parâmetros UTM** — configure `utm_source`, `utm_medium`, `utm_campaign`, `utm_term` e `utm_content` para cada link
- **Painel administrativo** — CRUD completo com login por senha única (bcrypt)
- **Contagem de visitas** — cada redirect registra o número de acessos
- **Ativar/desativar** — desative um redirect sem precisar excluí-lo
- **SQLite embutido** — sem dependência de banco externo, volume Docker persistente
- **Container mínimo** — imagem Docker ~7MB (Alpine + binário Go estático)
- **Tema escuro** — interface responsiva e agradável

## Stack


| Camada    | Tecnologia                                |
| --------- | ----------------------------------------- |
| Linguagem | Go 1.25                                   |
| Banco     | SQLite (WAL mode, via modernc.org/sqlite) |
| Frontend  | Go html/template + CSS (sem frameworks)   |
| Sessão    | Cookie-based em memória                   |
| Container | Docker multi-stage (Alpine)               |
| Porta     | 3000 (configurável via `PORT`)            |


## Início rápido

```bash
# 1. Clone e configure
git clone https://github.com/seu-usuario/go.raphai.eu.git
cd go.raphai.eu
cp .env.example .env
# Edite ADMIN_USER e ADMIN_PASS no .env

# 2. Inicie com Docker
docker compose up -d

# 3. Acesse
open http://localhost:3000/admin
```

## Configuração


| Variável     | Obrigatório | Padrão                | Descrição                 |
| ------------ | ----------- | --------------------- | ------------------------- |
| `ADMIN_USER` | Não         | `admin`               | Nome de usuário do painel |
| `ADMIN_PASS` | **Sim**     | —                     | Senha do painel           |
| `PORT`       | Não         | `3000`                | Porta do servidor HTTP    |
| `DB_PATH`    | Não         | `./data/redirects.db` | Caminho do arquivo SQLite |


## Como usar

### Painel admin

1. Acesse `http://localhost:3000/admin`
2. Faça login com as credenciais definidas no `.env`
3. Clique em **Novo** para criar um redirect
4. Preencha o **slug** (ex: `meu-projeto`) e a **URL de destino**
5. Escolha o **tipo de redirect**:
  - **Meta Refresh** — mostra uma tela de loading antes do redirect (SEO-friendly)
  - **301 Redirect** — redirect direto, sem loading
6. Adicione parâmetros **UTM** para rastreamento no Google Analytics
7. Ative/desative ou exclua redirects pela listagem

### Rotas


| Rota                | Descrição                           |
| ------------------- | ----------------------------------- |
| `/`                 | Redireciona para `/admin`           |
| `/:slug`            | Executa o redirect (loading ou 301) |
| `/admin`            | Dashboard do painel                 |
| `/admin/login`      | Login                               |
| `/admin/new`        | Criar novo redirect                 |
| `/admin/edit/:id`   | Editar redirect                     |
| `/admin/toggle/:id` | Ativar/desativar                    |
| `/admin/delete/:id` | Excluir redirect                    |


### Exemplo de UTM

Slug: `meu-projeto`  
Destino: `https://meuprojeto.com`  
UTM source: `raphai`  
UTM medium: `redirect`  
UTM campaign: `meu-projeto`

Resultado: `https://meuprojeto.com?utm_source=raphai&utm_medium=redirect&utm_campaign=meu-projeto`

## Deploy no Coolify

1. Crie um novo **Service** no Coolify
2. Conecte seu repositório git
3. Coolify detecta automaticamente o `docker-compose.yml`
4. Configure as variáveis de ambiente:
  - `ADMIN_USER`
  - `ADMIN_PASS`
5. Configure o domínio: `go.raphai.eu`
6. Defina a porta do container para **3000** (ou a configurada em `PORT`)
7. Deploy!

> O Coolify gerencia o proxy reverso (Traefik/Caddy) automaticamente. A porta do container é exposta internamente para o health check.

## Desenvolvimento

```bash
# Pré-requisitos: Go 1.22+
go version

# Instalar dependências
go mod tidy

# Executar em modo dev
DB_PATH=./data/redirects.db ADMIN_USER=admin ADMIN_PASS=dev123 go run .

# Build
go build -o server .
```

## Docker

```bash
# Build manual
docker build -t go-raphai-redirect .

# Executar com variáveis
docker run -d \
  --name go-raphai-redirect \
  -p 3000:3000 \
  -e ADMIN_USER=admin \
  -e ADMIN_PASS=suasenha \
  -v redirect_data:/data \
  go-raphai-redirect
```

## Projetos relacionados

- **[raphai.eu](https://raphai.eu)** — Currículo / portfólio pessoal

## Licença

MIT