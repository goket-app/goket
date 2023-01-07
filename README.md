# Goket

`Goket` (Golang Keyboard Event Tree) is a proof-of-concept code for using keyboard events trees for performing operations.

Its main goal is to allow using one or more keyboards as inputs for events and then translating one or more keypresses into actions.

This is mainly meant in IoT and DYI setups where a Linux-based device such as a Raspberry Pi along with a keyboard is used as a remote control for a device.

I am using it with multiple wireless numeric keyboards with USB receivers. I find it more convenient and easier to use than a smartphone with an application.

## Installation

Goket can be installed from the repository - such as:

```bash
go install github.com/goket-app/goket/cmd/goket
```

I'll also start providing pre-built binaries soon.

## Usage

Goket is started by running the `goket` command, optiionally providing `-config`, `-timeout` and `-devices` arguments:

```bash
goket -config ./path/to/config.yml -devices "/dev/input/event0,/dev/input/event1,..." -timeout 2.0
```

By default the configuration is read from `/etc/goket.json` and all input devices are read if the list is not specified. The default timeout is 2 seconds.

Where 

## Configuration

The configuration is a JSON file that contains a tree of key presses that should be processed - such as:

```json
{
  "keys": {
    "KEY_BACKSPACE": {
      "action": "http://local-server/control/all-lights/off"
    },
    "KEY_KP1": {
      "action": "http://local-server/control/main-light/toggle",
      "children": {
        "KEY_KPPLUS": {
          "action": "http://local-server/control/main-light/on"
        },
        "KEY_KPMINUS": {
          "action": "http://local-server/control/main-light/off"
        }
      }
    }
  }
}
```

The example above will suppor tthe following keypresses:

- pressing backspace key will cause the tool to invoke `http://local-server/control/all-lights/off` URL
- pressing keypad `1` without additional keys will cause the tool to invoke `http://local-server/control/main-light/toggle`
- pressing keypad `1`, followed by keypad `+` without additional keys will cause the tool to invoke `http://local-server/control/main-light/on`
- pressing keypad `1`, followed by keypad `-` without additional keys will cause the tool to invoke `http://local-server/control/main-light/off`

This allows defining single and multiple key combinations and mixing them.

The value for `-timeout` is the time after which `goket` will run actions for commands that have children defined, but none of the keys for child actions were pressed.
