## Appendix to the README
This documents new features added in this fork.

## Installing
A plain `go get...` won't work here because I decided not to change the module name (yet). Instead, clone and install "manually" like this:

```sh
$ git clone --depth=1 https://github.com/cqsd/gotty
$ cd gotty
$ GO11MODULE=on go install .
```

## Recording a session using Segment
> We've all gotta get our kicks somehow&mdash;I won't judge!

Track user input using [Segment track calls](https://segment.com/docs/connections/spec/track/) by providing a write key on the command line.

```sh
$ gotty --segment-write-key abcdefghijklmnopqrstuvwxyzABCDEF -w bash
```

Tracking events are sent server-side. The event (without the context set by the Segment SDK) looks about like this:
```go
client.Track(&analytics.Track{
  UserId: "[::1]:61232",
  Event: "gotty input",
  Properties: map[string]interface{}{
    "input": "\"echo hello  \u007gotty\"",
  },
})
```

The event name is "gotty input", and the user ID is set to the client's remote address. Unprintable inputs like `backspace`, `ctrl`, etc are quoted using Go's `%q`. You may also notice these byte sequences in the output:
 - `\x1b[O`, which indicates that the client has [unfocused the window](https://github.com/xtermjs/xterm.js/blob/8d912e891e367053d966733310e796c99ac99b60/src/browser/Terminal.ts#L278)
 - `\x1b[I`, which indicates that the client has [focused the window](https://github.com/xtermjs/xterm.js/blob/8d912e891e367053d966733310e796c99ac99b60/src/browser/Terminal.ts#L253)

These are sent by [xterm](https://github.com/xtermjs/xterm.js), an upstream dependency which provides the browser terminal interface.
