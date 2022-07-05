# Rebooter - an HTTP server to reboot on a new image

Rebooter is a small HTTP server written in Go whose purpose is to install a new
kernel image on the current machine and reboot on that image using GRUB. The
second next reboot will select the default image again, so that rebooter can be
reused without needing to SSH on the machine.

## Usage

On the server:

```sh
sudo ./rebooter <disk-device> <menu-entry>
# e.g.: sudo ./rebooter /dev/sdb myOS
```

And from a client:

```sh
curl http://<server-url>:8080/ --data-binary @<path-to-kernel-image>
```

## Prerequisites

- A GNU/Linux distribution, e.g. Ubuntu, with GRUB.
- A free disk or partition to use for the image.
- A GRUB menu entry corresponding to the target disk device must be created.

**Adding a new menu entry**:

Menu entries can be added by editing one of the files under `/etc/grub.d` (e.g.
`40_custom` in ubuntu), and then running `update-grub` to regenerate
`/boot/grub/grub.cfg`.

For instance, this will add a "myOS" menu entry booting on the second hard drive
(GRUB starts from `hd0`), i.e. `/dev/hdb` in Linux parlance.

```sh
menuentry "myOS" {
    set root="(hd1)"
    chainloader +1
}
```

## Start rebooter on Linux boot

With systemd:

Create a file `/etc/systemd/system/rebooter.service` and add the following
content:

```
[Unit]
Description=Start the rebooter HTTP server

[Service]
Type=oneshot
ExecStart=/bin/sh -c "/path/to/rebooter <disk-device> <menu-entry>"

[Install]
WantedBy=multi-user.target
```

Then enable the service

```sh
sudo systemctl enable --now rebooter.service
```

See [stack-overflow question](https://unix.stackexchange.com/a/47715).

## WARNING

Be **very** careful about which disk to target, this disk content will get
overwritten!

