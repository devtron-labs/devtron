# pp [![wercker status](https://app.wercker.com/status/6934c847631da2cf672e559f927a90b2/s "wercker status")](https://app.wercker.com/project/bykey/6934c847631da2cf672e559f927a90b2)

Colored pretty printer for Go language

![](http://i.gyazo.com/d3253ae839913b7239a7229caa4af551.png)

## Usage

Just call `pp.Print()`.

```go
import "github.com/k0kubun/pp"

m := map[string]string{"foo": "bar", "hello": "world"}
pp.Print(m)
```

![](http://i.gyazo.com/0d08376ed2656257627f79626d5e0cde.png)

### API

fmt package-like functions are provided.

```go
pp.Print()
pp.Println()
pp.Sprint()
pp.Fprintf()
// ...
```

API doc is available at: http://godoc.org/github.com/k0kubun/pp

### Custom colors

If you require, you may change the colors (all or some) for syntax highlighting:

```go
// Create a struct describing your scheme
scheme := pp.ColorScheme{
	Integer:       pp.Green | pp.Bold,
	Float:         pp.Black | pp.BackgroundWhite | pp.Bold,
	String:        pp.Yellow,
}

// Register it for usage
pp.SetColorScheme(scheme)
```

Look into ColorScheme struct for the field names.

If you would like to revert to the default highlighting, you may do so by calling `pp.ResetColorScheme()`.

Out of the following color flags, you may combine any color with a background color and optionally with the bold parameter. Please note that bold will likely not work on the windows platform.

```go
// Colors
Black
Red
Green
Yellow
Blue
Magenta
Cyan
White

// Background colors
BackgroundBlack
BackgroundRed
BackgroundGreen
BackgroundYellow
BackgroundBlue
BackgroundMagenta
BackgroundCyan
BackgroundWhite

// Other
Bold

// Special
NoColor
```

## Demo

### Timeline

![](http://i.gyazo.com/a8adaeec965db943486e35083cf707f2.png)

### UserStream event

![](http://i.gyazo.com/1e88915b3a6a9129f69fb5d961c4f079.png)

### Works on windows

![](http://i.gyazo.com/ab791997a980f1ab3ee2a01586efdce6.png)

## License

MIT License
