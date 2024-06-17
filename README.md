# native-app

This is the go native backend for the SAGE App.

## Building

To build the app, you need to have the go compiler installed. You can download
it from [here](https://golang.org/dl/).

On linux, you will also need to install
[GTK related libraries](https://github.com/webview/webview?tab=readme-ov-file#prerequisites).

Once you have the go compiler installed, you can run the following command to
build the app:

```bash
go build
```

> [!NOTE]  
> If you are having issues building, you should run `git tag` and checkout the
> latest tag. This will ensure that you are building the latest version of the
> app which is known to work.

You can also find pre-built binaries in the
[releases](https://github.com/sag-enhanced/native-app/releases) section.

## Issues

If you have any issues with the app, please open an issue in the
[issue tracker](https://github.com/sag-enhanced/sage-issues/issues). See
[SECURITY.md](SECURITY.md) for more information on how to report security
issues.

## LICENSE

This project is licensed under a SOURCE AVAILABLE license. See the
[LICENSE](LICENSE.md) file for more details.
