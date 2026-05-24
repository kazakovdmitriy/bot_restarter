# bot-restarter

TG-бот для перезагрузки VDS. Запускается прямо на сервере, выполняет `sudo reboot`.

## Команды

| Команда   | Описание |
|-----------|----------|
| `/start`  | Приветствие и список команд |
| `/status` | Uptime сервера |
| `/reboot` | Перезагрузка (с подтверждением) |

## Быстрая установка

Скачать бинарник из [релизов](https://github.com/kazakovdmitriy/bot_restarter/releases) и запустить:

```bash
# Создать пользователя
sudo useradd -r -s /bin/false botuser

# Разрешить reboot без пароля
echo 'botuser ALL=(ALL) NOPASSWD: /sbin/reboot' | sudo tee /etc/sudoers.d/bot-restarter

# Установить бота
sudo mkdir -p /opt/bot-restarter
sudo cp bot-restarter /opt/bot-restarter/
sudo cp config.json /opt/bot-restarter/
sudo chown -R botuser:botuser /opt/bot-restarter
sudo chmod +x /opt/bot-restarter/bot-restarter

# Установить systemd-сервис
sudo cp bot-restarter.service /etc/systemd/system/
sudo systemctl daemon-reload
sudo systemctl enable --now bot-restarter
```

## config.json

```json
{
  "telegram_token": "YOUR_BOT_TOKEN",
  "allowed_user_ids": [123456789]
}
```

## Сборка из исходников

```bash
go build -o bot-restarter .
```

Кросс-компиляция под Linux:

```bash
make build
```
