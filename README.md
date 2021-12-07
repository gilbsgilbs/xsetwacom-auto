# xsetwacom-auto

Simple CLI tool that uses
[xsetwacom](https://github.com/linuxwacom/xf86-input-wacom) and
[xrandr](https://gitlab.freedesktop.org/xorg/app/xrandr) to:

- map the to one monitor
- set the tablet area to match the aspect ratio of the monitor

> ℹ I wrote this project primarily for my personal setup and needs. It's not
> battled-tested and probably not very well written, but I'm publishing it in case
> it's useful to someone. I don't plane on working a lot on it.

## Usage

- `./xsetwacom-auto` will just assign all devices to the primary monitor.
- `./xsetwacom-auto --interactive` will show interactive CLI prompts to let you pick devices and a monitor.
- `./xsetwacom-auto --help` will show all options.
