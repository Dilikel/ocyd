# ocyd

Лёгкий демон для Linux, который передаёт текущую температуру CPU или GPU на дисплеи кулеров Ocypus по USB HID.

[![Лицензия: MIT](https://img.shields.io/badge/License-MIT-yellow.svg)](https://opensource.org/licenses/MIT)
[![Последний релиз](https://img.shields.io/github/v/release/Dilikel/ocyd?include_prereleases)](https://github.com/Dilikel/ocyd/releases)

[Read in English](README.md) · [Сообщить об ошибке или запросить поддержку](https://github.com/Dilikel/ocyd/issues/new/choose)

## Статус

ocyd находится на ранней альфа-стадии (`v0.1.0-alpha`), поэтому поддержка моделей ещё расширяется.

| Модель                | USB ID          | Статус                               |
| --------------------- | --------------- | ------------------------------------ |
| Ocypus Delta A62EX    | `0x1a2c:0x434d` | Протестирована                       |
| Другие дисплеи Ocypus | Неизвестно      | Пожалуйста, протестируйте и сообщите |

## Требования

- Linux с `systemd` и включёнными пользовательскими сервисами.
- Файлы разработки `libudev`. В Arch Linux: `sudo pacman -S systemd-libs`; в Ubuntu/Debian: `sudo apt install libudev-dev`.
- Go `1.26+` и `make` при сборке из исходников.

## Установка

```bash
git clone https://github.com/Dilikel/ocyd.git
cd ocyd
./install.sh
```

Инсталлятор соберёт демон, предложит выбрать профиль устройства, установит бинарник и udev-правила, создаст `~/.config/ocyd/config.toml` и включит пользовательский сервис. Для установки системных файлов может потребоваться `sudo`.

## Конфигурация

Файл по умолчанию находится в `~/.config/ocyd/config.toml`:

```toml
[display]
source = "cpu" # "cpu" или "gpu"
unit = "C"     # "C" или "F"

[device]
vendor_id = 0x1a2c
product_id = 0x434d
```

Чтобы использовать другой дисплей, измените уже установленные конфигурацию и udev-правило, как описано ниже. Демон читает конфигурацию при запуске.

## Сервис и логи

```bash
systemctl --user status ocyd.service
journalctl --user -u ocyd.service -f
```

Для остановки или перезапуска используйте `systemctl --user stop ocyd.service` или `systemctl --user restart ocyd.service`.

## Тестирование своего кулера

Подключите кулер и найдите его USB ID:

```bash
lsusb
```

В выводе кулер может определиться как клавиатура. Например, Delta A62EX отображается так:

```text
Bus 001 Device 005: ID 1a2c:434d China Resource Semico Co., Ltd USB Gaming Keyboard
```

Другие кулеры также могут распознаваться как клавиатура. В этой строке `1a2c` — это **vendor ID**, а `434d` — **product ID**. Используйте ID, которые показывает `lsusb` для вашего устройства.

После того как вы нашли свои ID, измените эти два файла:

1. Конфигурация устройства: `~/.config/ocyd/config.toml`

   ```toml
   [device]
   vendor_id = 0x1234
   product_id = 0x5678
   ```

2. Разрешения udev: `/etc/udev/rules.d/99-ocypus.rules`

   ```text
   SUBSYSTEM=="hidraw", ATTRS{idVendor}=="1234", ATTRS{idProduct}=="5678", MODE="0666"
   ```

Замените `1234` и `5678` на vendor и product ID вашего устройства **в обоих файлах**. В udev-правиле оставьте формат из `lsusb`: четыре шестнадцатеричных символа без префикса `0x`. В TOML-конфигурации используется префикс `0x`.

Изменение udev-правила требует `sudo`, потому что файл принадлежит root.

После изменения udev-правила перезагрузите его и переподключите кулер:

```bash
sudo udevadm control --reload-rules
sudo udevadm trigger
```

Переподключите кулер, перезапустите демон и проверьте логи:

```bash
systemctl --user restart ocyd.service
journalctl --user -u ocyd.service -f
```

## Тестирование других моделей

Пока подтверждена только работа с Delta A62EX. Если у вас есть другая модель Ocypus, пожалуйста, протестируйте её и создайте [issue](https://github.com/Dilikel/ocyd/issues/new/choose), даже если устройство пока не работает. Укажите:

- точную модель кулера и дистрибутив Linux;
- USB vendor/product ID из `lsusb`;
- определяется ли дисплей и обновляется ли температура;
- важный вывод команд `systemctl --user status ocyd.service` и `journalctl --user -u ocyd.service`.

Профили устройств и pull request приветствуются. Не прикладывайте персональные данные и полные логи, не связанные с устройством.

## Сборка и разработка

```bash
make build
make test
make all
```

`make all` форматирует, проверяет, запускает линтеры и тесты, а затем собирает проект.

## План развития

- Добавить профили большего числа протестированных устройств.
- Расширить юнит-тесты и тестирование на виртуальном hardware tree.
- Добавить `ocydctl` для удобного управления демоном.
- Опубликовать пакет в AUR.

## Лицензия

ocyd распространяется по [лицензии MIT](LICENSE).
