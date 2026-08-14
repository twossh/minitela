# MiniTela

Aplicativo Linux nativo para gerenciamento da tela auxiliar do **Positivo Vision R15M**.

O projeto está em desenvolvimento ativo e tem como foco controlar a MiniTela do R15M diretamente no Linux, sem depender do software original do fabricante.

## Status atual

Estado: **experimental / desenvolvimento**.

Validado fisicamente em:

- Positivo Vision R15M;
- controlador GC9002;
- USB CDC ACM `0324:0324`;
- Fedora KDE;
- arquitetura `x86_64`.

Ainda não há uma release pública publicada no GitHub.

## Funcionalidades atuais

- Interface gráfica em Go + Fyne;
- detecção automática da MiniTela;
- reconexão automática do dispositivo;
- ligar, desligar e reiniciar o controlador;
- seleção de tela: Monitor, Clima, Notas, WhatsApp e Imagem;
- configuração de cidade para o clima;
- controle de brilho;
- restauração da última tela utilizada;
- backend persistente via `systemd --user`;
- opção de iniciar automaticamente com o sistema;
- diagnóstico do ambiente;
- configuração automática da regra `udev` no primeiro uso;
- alias estável `/dev/minitela-r15m`;
- upload nativo de imagens personalizadas;
- suporte validado para JPG, PNG e GIF animado;
- distribuição em AppImage.

## Pipeline de imagem personalizado

O pipeline atualmente validado é:

```text
JPG / PNG / GIF
        ↓
decode e composição em Go
        ↓
192 × 192
        ↓
STCRGBA nativo
        ↓
ACF de 30 frames
        ↓
upload nativo para o R15M
        ↓
GC9002
```

JPG, PNG e GIF animado foram testados fisicamente no hardware.

O AppImage não precisa de uma galeria de imagens do fabricante para que a função **Minha imagem** funcione.

## AppImage autossuficiente

O fluxo de imagem personalizada foi reorganizado para que o usuário final não precise conhecer ou importar arquivos ACF.

O AppImage contém apenas os componentes necessários ao MiniTela, como:

```text
usr/share/minitela/
├── systemd/minitela.service
├── template/Texture-template.acf
└── udev/99-minitela.rules
```

A pasta `vendor/`, previews, GIFs e imagens da galeria original não fazem parte do empacotamento atual.

### Observação importante sobre o template ACF

O template neutro atualmente usado no AppImage foi validado fisicamente, porém ainda depende de um envelope ACF estudado durante a engenharia reversa.

Por esse motivo:

- arquivos `*.acf` não são versionados no repositório;
- assets originais do fabricante não são versionados;
- a criação de uma release pública permanece bloqueada até que o envelope ACF possa ser gerado integralmente pelo próprio MiniTela.

O próximo marco do projeto é eliminar essa última dependência de desenvolvimento com um gerador ACF próprio.

## Build do AppImage

O build atual suporta `x86_64`.

Antes de gerar o AppImage, é necessário apontar explicitamente para o template neutro usado no ambiente de desenvolvimento:

```bash
MINITELA_TEMPLATE_FILE=/caminho/Texture-neutral.acf \
  ./packaging/appimage/build-appimage.sh
```

O arquivo é incorporado ao AppImage durante o build e não precisa existir no computador do usuário final.

Os artefatos gerados ficam em:

```text
dist/
```

## Testes

Para executar a suíte de testes:

```bash
go test ./...
```

Para validar os scripts de empacotamento:

```bash
bash -n packaging/appimage/AppRun
bash -n packaging/appimage/build-appimage.sh
```

## Segurança do repositório

O projeto ignora deliberadamente:

```text
/assets/vendor/
*.acf
/dist/
```

Isso reduz o risco de publicação acidental de arquivos usados apenas durante a pesquisa e os testes locais.

## Estrutura principal

```text
cmd/
  minitela-gui/       Interface gráfica
  minitela/           Backend
  minitela-ctl/       Controle do dispositivo

internal/
  acf/                Estrutura e checksum ACF
  customimage/        Pipeline de imagens personalizadas
  device/             Detecção do hardware
  gallery/            Suporte de galeria opcional
  protocol/           Protocolo de comunicação
  r15m/               Comunicação específica com o R15M
  stc/                Encoder STCRGBA
  textureupload/      Upload e ativação de textura

packaging/
  appimage/           Empacotamento AppImage
  udev/               Regra de dispositivo
```

## Próximos passos

1. mapear completamente o envelope ACF de 30 frames;
2. gerar o ACF integralmente em código, sem template externo de desenvolvimento;
3. adicionar testes automatizados do novo gerador;
4. validar novamente JPG, PNG e GIF no hardware;
5. ampliar a validação para outras distribuições Linux;
6. preparar a primeira release pública.

## Versão

Versão de desenvolvimento: `0.1.0-dev`
