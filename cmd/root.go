package cmd

import (
	"errors"
	"fmt"
	"net/http"
	"os"

	"github.com/apex/log"
	"github.com/gedsonn/zaapi/config"
	"github.com/gedsonn/zaapi/maneger"
	server "github.com/gedsonn/zaapi/server/http"
	"github.com/spf13/cobra"
)

var (
	configPath string
	debug      bool
)

var rootCmd = &cobra.Command{
	Use:     "zaapi",
	Short:   "Zaapi - Serviço API",
	Long:    `Zaapi é um serviço API modular com suporte a sessões, WhatsApp e WebSockets.`,
	Version: "1.0.0",

	// Executado antes de qualquer comando
	PersistentPreRun: func(cmd *cobra.Command, args []string) {
		if debug {
			log.Info("RODANDO EM MODO DE DEBUG")
			os.Setenv("ZAAPI_DEBUG", "true")
		}
	},

	// Executado quando o usuário digita somente "zaapi"
	Run: start,
}

func start(cmd *cobra.Command, args []string) {
	cwd, err := os.Getwd()
	if err != nil {
		log.Warn("Falha ao obter diretório atual, usando fallback em /etc/zaapi/config.yml")
	}

	configPath = fmt.Sprintf("%s/config.yml", cwd)
	log.Infof("Usando arquivo de configuração: %s", configPath)

	//Carregar aquivo de configuração
	err = config.FromFile(configPath)
	if err != nil {
		//validar se o arquivo existe
		if errors.Is(err, os.ErrNotExist) {

			// arquivo não existe → criar novo config

			cfg, _ := config.NewAtPath(configPath)
			config.Set(cfg)
			// salvar novo config no arquivo
			if err := config.Save(configPath); err != nil {
				log.Fatalf("Falha ao salvar novo arquivo de configuração: %s", err)
			}

			log.Infof("Novo arquivo de configuração criado com sucesso.")
		} else {
			log.Fatalf("Erro ao tentar ler o arquivo de configuração: %s", err)
		}
	}

	//pegar configuração atual
	cfg := config.Get()
	m, err := manager.NewManager()
	if err != nil {
		panic(err)
	}

	if cfg.Server.Enable {
		//inicializar o servidor http
		log.Infof("Iniciando servidor HTTP na porta %d", cfg.Server.Port)
		go func() {
			s := server.Configure(m)
			if err := http.ListenAndServe(fmt.Sprintf("%s:%d", cfg.Server.Host, cfg.Server.Port), s); err != nil {
				log.Fatalf("Erro ao iniciar o servidor HTTP: %v", err)
			}
		}()
	}

	//impedir que o programa termine
	select {}
}

var versionCommand = &cobra.Command{
	Use:   "version",
	Short: "Mostra a versão atual do Zaapi",
	Run: func(cmd *cobra.Command, _ []string) {
		fmt.Println("Zaapi version 1.0.0")
	},
}

var debugCommand = &cobra.Command{
	Use:   "debug",
	Short: "Habilita o modo de depuração",
	Run: func(cmd *cobra.Command, _ []string) {
		debug = true
		fmt.Println("Debug mode enabled")
	},
}

func Execute() {
	if err := rootCmd.Execute(); err != nil {
		log.Fatalf("Erro ao executar o Zaapi: %v", err)
	}
}

func init() {
	rootCmd.AddCommand(versionCommand)
	rootCmd.AddCommand(debugCommand)

	rootCmd.PersistentFlags().BoolVarP(
		&debug,
		"debug",
		"d",
		false,
		"Ativar modo de depuração",
	)
}
