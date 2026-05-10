# Instalação de APKs no Android via Termux

Este guia explica como instalar aplicativos Android (arquivos .apk) usando o Termux, com e sem acesso root.

## Pré-requisitos

Antes de começar, certifique-se de que o Termux tem as ferramentas necessárias instaladas:

```bash
# Atualizar o Termux
pkg update && pkg upgrade

# Instalar as ferramentas do Android (ADB e Fastboot)
pkg install android-tools
```

---

## Método 1: Usando ADB (Sem Root)

Este método usa o Android Debug Bridge (ADB) via Wi-Fi para instalar aplicativos. É o método mais seguro e não requer root.

### Passo 1: Preparar o Dispositivo

1. Vá em **Configurações > Opções do Desenvolvedor**
2. Ative a **Depuração USB**
3. Ative a **Depuração por Wi-Fi** (pode aparecer como "Depuração sem fio" ou "Wireless Debugging")
4. Toque em "Depuração por Wi-Fi" e depois em **"Parear dispositivo com código de pareamento"**
5. **Anote** as seguintes informações:
   - Endereço IP
   - Porta de pareamento
   - Código de pareamento

### Passo 2: Conectar o Termux ao ADB

No terminal do Termux, execute:

```bash
# Parear o dispositivo (substitua pelos dados anotados)
adb pair [IP]:[Porta_do_Pareamento] [Código_de_Pareamento]

# Conectar ao dispositivo (use o IP e Porta principal da tela de Depuração Wi-Fi)
adb connect [IP]:[Porta_Principal]

# Verificar se a conexão foi estabelecida
adb devices
```

Deve aparecer seu dispositivo na lista de dispositivos conectados.

### Passo 3: Instalar o Aplicativo

Com o dispositivo conectado e o arquivo APK localizado (por exemplo, na pasta `/sdcard/Download/`):

```bash
# Instalação padrão
adb install /sdcard/Download/nome-do-seu-app.apk

# Se quiser substituir um app existente mantendo seus dados
adb install -r /sdcard/Download/nome-do-seu-app.apk
```

**Nota:** O caminho do APK pode ser qualquer localização acessível no dispositivo. Certifique-se de que o arquivo existe e o caminho está correto.

---

## Método 2: Usando Root (pm e am)

Se o seu dispositivo já possui acesso root, você pode instalar aplicativos diretamente usando o gerenciador de pacotes nativo do Android (`pm`) sem precisar configurar o ADB.

Este método é mais rápido e não requer conexão de rede.

### Passo 1: Obter Acesso Root no Termux

```bash
su
```

O gerenciador de root (Magisk, Kitsune, ou outro) solicitará permissão. Aceite a solicitação.

### Passo 2: Instalar usando Package Manager (pm)

```bash
pm install /sdcard/Download/nome-do-seu-app.apk
```

O comando `pm install` é a forma direta e nativa de instalar um APK com privilégios de root.

### Passo 3: Gerenciar o App após Instalação (Activity Manager)

O comando `am` (Activity Manager) é usado para gerenciar o aplicativo após a instalação, como forçar a abertura de uma activity específica ou iniciar um serviço.

Exemplo: Abrir o app logo após instalá-lo:

```bash
am start -n com.ask.myapp/.MainActivity
```

**Substitua:**
- `com.ask.myapp` pelo package name correto do aplicativo
- `.MainActivity` pela activity principal do app

---

## Resumo de Comandos

| Situação | Comando Principal | Requisito |
|----------|-------------------|-----------|
| Instalação Padrão (sem root) | `adb install app.apk` | Depuração Wi-Fi ativada e conectada |
| Substituir App Existente (sem root) | `adb install -r app.apk` | Mantém os dados do app antigo |
| Instalação com Root | `pm install app.apk` | Permissão de Superusuário (`su`) no Termux |
| Iniciar App (com root) | `am start -n pacote/activity` | App já instalado com root |

---

## Solução de Problemas

### Erro: "error: closed" ou "error: device offline"
- Verifique se a depuração Wi-Fi está ativa
- Tente reconectar: `adb disconnect` e depois `adb connect` novamente
- Confirme se o IP e porta estão corretos

### Erro: "Failure [INSTALL_FAILED_ALREADY_EXISTS]"
- Use a flag `-r` para reinstalar mantendo dados: `adb install -r app.apk`
- Ou desinstale primeiro: `adb uninstall nome.do.pacote`

### Erro: "Failure [INSTALL_PARSE_FAILED_NO_CERTIFICATES]"
- O APK não está assinado ou a assinatura é inválida
- Certifique-se de que o APK é de uma fonte confiável

### Erro de permissão com `pm install` ou `am start`
- Verifique se você executou `su` e concedeu permissão de root
- Alguns dispositivos com root podem exigir configurações adicionais

---

## Dicas de Segurança

- **Sempre verifique a fonte do APK**: Instale apenas aplicativos de fontes confiáveis
- **Revise permissões**: Após instalar, verifique as permissões concedidas ao app
- **Root**: Ter root no dispositivo pode comprometer a segurança e anular garantias
- **Use ADB sem root quando possível**: É mais seguro e não modifica o sistema

---

## Referências

- [Documentação oficial do Android Debug Bridge](https://developer.android.com/studio/command-line/adb)
- [Android Package Manager (pm)](https://developer.android.com/studio/command-line/adb#pm)
- [Android Activity Manager (am)](https://developer.android.com/studio/command-line/adb#am)
