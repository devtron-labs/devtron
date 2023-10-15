# mo - Monads

[![tag](https://img.shields.io/github/tag/samber/mo.svg)](https://github.com/samber/mo/releases)
![Go Version](https://img.shields.io/badge/Go-%3E%3D%201.18-%23007d9c)
[![GoDoc](https://godoc.org/github.com/samber/mo?status.svg)](https://pkg.go.dev/github.com/samber/mo)
![Build Status](https://github.com/samber/mo/actions/workflows/test.yml/badge.svg)
[![Go report](https://goreportcard.com/badge/github.com/samber/mo)](https://goreportcard.com/report/github.com/samber/mo)
[![Coverage](https://img.shields.io/codecov/c/github/samber/do)](https://codecov.io/gh/samber/mo)
[![License](https://img.shields.io/github/license/samber/mo)](./LICENSE)

🦄 **`samber/mo` brings monads and popular FP abstractions to Go projects. `samber/mo` uses the recent Go 1.18+ Generics.**

**Inspired by:**

- Scala
- Rust
- FP-TS

**See also:**

- [samber/lo](https://github.com/samber/lo): A Lodash-style Go library based on Go 1.18+ Generics
- [samber/do](https://github.com/samber/do): A dependency injection toolkit based on Go 1.18+ Generics

**Why this name?**

I love **short name** for such utility library. This name is similar to "Monad Go" and no Go package currently uses this name.

## 💡 Features

We currently support the following data types:

- `Option[T]` (Maybe)
- `Result[T]`
- `Either[A, B]`
- `EitherX[T1, ..., TX]` (With X between 3 and 5)
- `Future[T]`
- `IO[T]`
- `IOEither[T]`
- `Task[T]`
- `TaskEither[T]`
- `State[S, A]`

## 🚀 Install

```sh
go get github.com/samber/mo@v1
```

This library is v1 and follows SemVer strictly.

No breaking changes will be made to exported APIs before v2.0.0.

This library has no dependencies except the Go std lib.

## 💡 Quick start

You can import `mo` using:

```go
import (
    "github.com/samber/mo"
)
```

Then use one of the helpers below:

```go
option1 := mo.Some(42)
// Some(42)

option1.
    FlatMap(func (value int) Option[int] {
        return Some(value*2)
    }).
    FlatMap(func (value int) Option[int] {
        return Some(value%2)
    }).
    FlatMap(func (value int) Option[int] {
        return Some(value+21)
    }).
    OrElse(1234)
// 21

option2 := mo.None[int]()
// None

option2.OrElse(1234)
// 1234

option3 := option1.Match(
    func(i int) (int, bool) {
        // when value is present
        return i * 2, true
    },
    func() (int, bool) {
        // when value is absent
        return 0, false
    },
)
// Some(42)
```

More examples in [documentation](https://godoc.org/github.com/samber/mo).

### Tips for lazy developers

I cannot recommend it, but in case you are too lazy for repeating `mo.` everywhere, you can import the entire library into the namespace.

```go
import (
    . "github.com/samber/mo"
)
```

I take no responsibility on this junk. 😁 💩

## 🤠 Documentation and examples

[GoDoc: https://godoc.org/github.com/samber/mo](https://godoc.org/github.com/samber/mo)

### Option[T any]

`Option` is a container for an optional value of type `T`. If value exists, `Option` is of type `Some`. If the value is absent, `Option` is of type `None`.

Constructors:

- `mo.Some()` [doc](https://pkg.go.dev/github.com/samber/mo#Some) - [play](https://go.dev/play/p/iqz2n9n0tDM)
- `mo.None()` [doc](https://pkg.go.dev/github.com/samber/mo#None) - [play](https://go.dev/play/p/yYQPsYCSYlD)
- `mo.TupleToOption()` [doc](https://pkg.go.dev/github.com/samber/mo#TupleToOption) - [play](https://go.dev/play/p/gkrg2pZwOty)
- `mo.EmptyableToOption()` [doc](https://pkg.go.dev/github.com/samber/mo#EmptyableToOption) - [play](https://go.dev/play/p/GSpQQ-q-UES)
- `mo.PointerToOption()` [doc](https://pkg.go.dev/github.com/samber/mo#PointerToOption) - [play](https://go.dev/play/p/yPVMj4DUb-I)

Methods:

- `.IsPresent()` [doc](https://pkg.go.dev/github.com/samber/mo#Option.IsPresent) - [play](https://go.dev/play/p/nDqIaiihyCA)
- `.IsAbsent()` [doc](https://pkg.go.dev/github.com/samber/mo#Option.IsAbsent) - [play](https://go.dev/play/p/23e2zqyVOQm)
- `.Size()` [doc](https://pkg.go.dev/github.com/samber/mo#Option.Size) - [play](https://go.dev/play/p/7ixCNG1E9l7)
- `.Get()` [doc](https://pkg.go.dev/github.com/samber/mo#Option.Get) - [play](https://go.dev/play/p/0-JBa1usZRT)
- `.MustGet()` [doc](https://pkg.go.dev/github.com/samber/mo#Option.MustGet) - [play](https://go.dev/play/p/RVBckjdi5WR)
- `.OrElse()` [doc](https://pkg.go.dev/github.com/samber/mo#Option.OrElse) - [play](https://go.dev/play/p/TrGByFWCzXS)
- `.OrEmpty()` [doc](https://pkg.go.dev/github.com/samber/mo#Option.OrEmpty) - [play](https://go.dev/play/p/SpSUJcE-tQm)
- `.ToPointer()` [doc](https://pkg.go.dev/github.com/samber/mo#Option.ToPointer) - [play](https://go.dev/play/p/N43w92SM-Bs)
- `.ForEach()` [doc](https://pkg.go.dev/github.com/samber/mo#Option.ForEach)
- `.Match()` [doc](https://pkg.go.dev/github.com/samber/mo#Option.Match) - [play](https://go.dev/play/p/1V6st3LDJsM)
- `.Map()` [doc](https://pkg.go.dev/github.com/samber/mo#Option.Map) - [play](https://go.dev/play/p/mvfP3pcP_eJ)
- `.MapNone()` [doc](https://pkg.go.dev/github.com/samber/mo#Option.MapNone) - [play](https://go.dev/play/p/_KaHWZ6Q17b)
- `.FlatMap()` [doc](https://pkg.go.dev/github.com/samber/mo#Option.FlatMap) - [play](https://go.dev/play/p/OXO-zJx6n5r)
- `.MarshalJSON()` [doc](https://pkg.go.dev/github.com/samber/mo#Option.MarshalJSON)
- `.UnmarshalJSON()` [doc](https://pkg.go.dev/github.com/samber/mo#Option.UnmarshalJSON)
- `.MarshalText()` [doc](https://pkg.go.dev/github.com/samber/mo#Option.MarshalText)
- `.UnmarshalText()` [doc](https://pkg.go.dev/github.com/samber/mo#Option.UnmarshalText)
- `.MarshalBinary()` [doc](https://pkg.go.dev/github.com/samber/mo#Option.MarshalBinary)
- `.UnmarshalBinary()` [doc](https://pkg.go.dev/github.com/samber/mo#Option.UnmarshalBinary)
- `.GobEncode()` [doc](https://pkg.go.dev/github.com/samber/mo#Option.GobEncode)
- `.GobDecode()` [doc](https://pkg.go.dev/github.com/samber/mo#Option.GobDecode)
- `.Scan()` [doc](https://pkg.go.dev/github.com/samber/mo#Option.Scan)
- `.Value()` [doc](https://pkg.go.dev/github.com/samber/mo#Option.Value)

### Result[T any]

`Result` respresent a result of an action having one of the following output: success or failure. An instance of `Result` is an instance of either `Ok` or `Err`. It could be compared to `Either[error, T]`.

Constructors:

- `mo.Ok()` [doc](https://pkg.go.dev/github.com/samber/mo#Ok) - [play](https://go.dev/play/p/PDwADdzNoyZ)
- `mo.Err()` [doc](https://pkg.go.dev/github.com/samber/mo#Err) - [play](https://go.dev/play/p/PDwADdzNoyZ)
- `mo.Errf()` [doc](https://pkg.go.dev/github.com/samber/mo#Errf) - [play](https://go.dev/play/p/N43w92SM-Bs)
- `mo.TupleToResult()` [doc](https://pkg.go.dev/github.com/samber/mo#TupleToResult) - [play](https://go.dev/play/p/KWjfqQDHQwa)
- `mo.Try()` [doc](https://pkg.go.dev/github.com/samber/mo#Try) - [play](https://go.dev/play/p/ilOlQx-Mx42)

Methods:

- `.IsOk()` [doc](https://pkg.go.dev/github.com/samber/mo#Result.IsOk) - [play](https://go.dev/play/p/sfNvBQyZfgU)
- `.IsError()` [doc](https://pkg.go.dev/github.com/samber/mo#Result.IsError) - [play](https://go.dev/play/p/xkV9d464scV)
- `.Error()` [doc](https://pkg.go.dev/github.com/samber/mo#Result.Error) - [play](https://go.dev/play/p/CSkHGTyiXJ5)
- `.Get()` [doc](https://pkg.go.dev/github.com/samber/mo#Result.Get) - [play](https://go.dev/play/p/8KyX3z6TuNo)
- `.MustGet()` [doc](https://pkg.go.dev/github.com/samber/mo#Result.MustGet) - [play](https://go.dev/play/p/8LSlndHoTAE)
- `.OrElse()` [doc](https://pkg.go.dev/github.com/samber/mo#Result.OrElse) - [play](https://go.dev/play/p/MN_ULx0soi6)
- `.OrEmpty()` [doc](https://pkg.go.dev/github.com/samber/mo#Result.OrEmpty) - [play](https://go.dev/play/p/rdKtBmOcMLh)
- `.ToEither()` [doc](https://pkg.go.dev/github.com/samber/mo#Result.ToEither) - [play](https://go.dev/play/p/Uw1Zz6b952q)
- `.ForEach()` [doc](https://pkg.go.dev/github.com/samber/mo#Result.ForEach)
- `.Match()` [doc](https://pkg.go.dev/github.com/samber/mo#Result.Match) - [play](https://go.dev/play/p/-_eFaLJ31co)
- `.Map()` [doc](https://pkg.go.dev/github.com/samber/mo#Result.Map) - [play](https://go.dev/play/p/-ndpN_b_OSc)
- `.MapErr()` [doc](https://pkg.go.dev/github.com/samber/mo#Result.MapErr) - [play](https://go.dev/play/p/WraZixg9GGf)
- `.FlatMap()` [doc](https://pkg.go.dev/github.com/samber/mo#Result.FlatMap) - [play](https://go.dev/play/p/Ud5QjZOqg-7)

### Either[L any, R any]

`Either` respresents a value of 2 possible types. An instance of `Either` is an instance of either `A` or `B`.

Constructors:

- `mo.Left()` [doc](https://pkg.go.dev/github.com/samber/mo#Left)
- `mo.Right()` [doc](https://pkg.go.dev/github.com/samber/mo#Right)

Methods:

- `.IsLeft()` [doc](https://pkg.go.dev/github.com/samber/mo#Either.IsLeft)
- `.IsRight()` [doc](https://pkg.go.dev/github.com/samber/mo#Either.IsRight)
- `.Left()` [doc](https://pkg.go.dev/github.com/samber/mo#Either.Left)
- `.Right()` [doc](https://pkg.go.dev/github.com/samber/mo#Either.Right)
- `.MustLeft()` [doc](https://pkg.go.dev/github.com/samber/mo#Either.MustLeft)
- `.MustRight()` [doc](https://pkg.go.dev/github.com/samber/mo#Either.MustRight)
- `.Unpack()` [doc](https://pkg.go.dev/github.com/samber/mo#Either.Unpack)
- `.LeftOrElse()` [doc](https://pkg.go.dev/github.com/samber/mo#Either.LeftOrElse)
- `.RightOrElse()` [doc](https://pkg.go.dev/github.com/samber/mo#Either.RightOrElse)
- `.LeftOrEmpty()` [doc](https://pkg.go.dev/github.com/samber/mo#Either.LeftOrEmpty)
- `.RightOrEmpty()` [doc](https://pkg.go.dev/github.com/samber/mo#Either.RightOrEmpty)
- `.Swap()` [doc](https://pkg.go.dev/github.com/samber/mo#Either.Swap)
- `.ForEach()` [doc](https://pkg.go.dev/github.com/samber/mo#Either.ForEach)
- `.Match()` [doc](https://pkg.go.dev/github.com/samber/mo#Either.Match)
- `.MapLeft()` [doc](https://pkg.go.dev/github.com/samber/mo#Either.MapLeft)
- `.MapRight()` [doc](https://pkg.go.dev/github.com/samber/mo#Either.MapRight)

### EitherX[T1, ..., TX] (With X between 3 and 5)

`EitherX` respresents a value of X possible types. For example, an `Either3` value is either `T1`, `T2` or `T3`.

Constructors:

- `mo.NewEitherXArgY()` [doc](https://pkg.go.dev/github.com/samber/mo#NewEither5Arg1). Eg:
  - `mo.NewEither3Arg1[A, B, C](A)`
  - `mo.NewEither3Arg2[A, B, C](B)`
  - `mo.NewEither3Arg3[A, B, C](C)`
  - `mo.NewEither4Arg1[A, B, C, D](A)`
  - `mo.NewEither4Arg2[A, B, C, D](B)`
  - ...

Methods:

- `.IsArgX()` [doc](https://pkg.go.dev/github.com/samber/mo#Either5.IsArg1)
- `.ArgX()` [doc](https://pkg.go.dev/github.com/samber/mo#Either5.Arg1)
- `.MustArgX()` [doc](https://pkg.go.dev/github.com/samber/mo#Either5.MustArg1)
- `.Unpack()` [doc](https://pkg.go.dev/github.com/samber/mo#Either5.Unpack)
- `.ArgXOrElse()` [doc](https://pkg.go.dev/github.com/samber/mo#Either5.Arg1OrElse)
- `.ArgXOrEmpty()` [doc](https://pkg.go.dev/github.com/samber/mo#Either5.Arg1OrEmpty)
- `.ForEach()` [doc](https://pkg.go.dev/github.com/samber/mo#Either5.ForEach)
- `.Match()` [doc](https://pkg.go.dev/github.com/samber/mo#Either5.Match)
- `.MapArgX()` [doc](https://pkg.go.dev/github.com/samber/mo#Either5.MapArg1)

### Future[T any]

`Future` represents a value which may or may not currently be available, but will be available at some point, or an exception if that value could not be made available.

Constructors:

- `mo.NewFuture()` [doc](https://pkg.go.dev/github.com/samber/mo#NewFuture)

Methods:

- `.Then()` [doc](https://pkg.go.dev/github.com/samber/mo#Future.Then)
- `.Catch()` [doc](https://pkg.go.dev/github.com/samber/mo#Future.Catch)
- `.Finally()` [doc](https://pkg.go.dev/github.com/samber/mo#Future.Finally)
- `.Collect()` [doc](https://pkg.go.dev/github.com/samber/mo#Future.Collect)
- `.Result()` [doc](https://pkg.go.dev/github.com/samber/mo#Future.Result)
- `.Cancel()` [doc](https://pkg.go.dev/github.com/samber/mo#Future.Cancel)

### IO[T any]

`IO` represents a non-deterministic synchronous computation that can cause side effects, yields a value of type `R` and never fails.

Constructors:

- `mo.NewIO()` [doc](https://pkg.go.dev/github.com/samber/mo#NewIO)
- `mo.NewIO1()` [doc](https://pkg.go.dev/github.com/samber/mo#NewIO1)
- `mo.NewIO2()` [doc](https://pkg.go.dev/github.com/samber/mo#NewIO2)
- `mo.NewIO3()` [doc](https://pkg.go.dev/github.com/samber/mo#NewIO3)
- `mo.NewIO4()` [doc](https://pkg.go.dev/github.com/samber/mo#NewIO4)
- `mo.NewIO5()` [doc](https://pkg.go.dev/github.com/samber/mo#NewIO5)

Methods:

- `.Run()` [doc](https://pkg.go.dev/github.com/samber/mo#Future.Run)

### IOEither[T any]

`IO` represents a non-deterministic synchronous computation that can cause side effects, yields a value of type `R` and can fail.

Constructors:

- `mo.NewIOEither()` [doc](https://pkg.go.dev/github.com/samber/mo#NewIOEither)
- `mo.NewIOEither1()` [doc](https://pkg.go.dev/github.com/samber/mo#NewIOEither1)
- `mo.NewIOEither2()` [doc](https://pkg.go.dev/github.com/samber/mo#NewIOEither2)
- `mo.NewIOEither3()` [doc](https://pkg.go.dev/github.com/samber/mo#NewIOEither3)
- `mo.NewIOEither4()` [doc](https://pkg.go.dev/github.com/samber/mo#NewIOEither4)
- `mo.NewIOEither5()` [doc](https://pkg.go.dev/github.com/samber/mo#NewIOEither5)

Methods:

- `.Run()` [doc](https://pkg.go.dev/github.com/samber/mo#IOEither.Run)

### Task[T any]

`Task` represents a non-deterministic asynchronous computation that can cause side effects, yields a value of type `R` and never fails.

Constructors:

- `mo.NewTask()` [doc](https://pkg.go.dev/github.com/samber/mo#NewTask)
- `mo.NewTask1()` [doc](https://pkg.go.dev/github.com/samber/mo#NewTask1)
- `mo.NewTask2()` [doc](https://pkg.go.dev/github.com/samber/mo#NewTask2)
- `mo.NewTask3()` [doc](https://pkg.go.dev/github.com/samber/mo#NewTask3)
- `mo.NewTask4()` [doc](https://pkg.go.dev/github.com/samber/mo#NewTask4)
- `mo.NewTask5()` [doc](https://pkg.go.dev/github.com/samber/mo#NewTask5)
- `mo.NewTaskFromIO()` [doc](https://pkg.go.dev/github.com/samber/mo#NewTaskFromIO)
- `mo.NewTaskFromIO1()` [doc](https://pkg.go.dev/github.com/samber/mo#NewTaskFromIO1)
- `mo.NewTaskFromIO2()` [doc](https://pkg.go.dev/github.com/samber/mo#NewTaskFromIO2)
- `mo.NewTaskFromIO3()` [doc](https://pkg.go.dev/github.com/samber/mo#NewTaskFromIO3)
- `mo.NewTaskFromIO4()` [doc](https://pkg.go.dev/github.com/samber/mo#NewTaskFromIO4)
- `mo.NewTaskFromIO5()` [doc](https://pkg.go.dev/github.com/samber/mo#NewTaskFromIO5)

Methods:

- `.Run()` [doc](https://pkg.go.dev/github.com/samber/mo#Task.Run)

### TaskEither[T any]

`TaskEither` represents a non-deterministic asynchronous computation that can cause side effects, yields a value of type `R` and can fail.

Constructors:

- `mo.NewTaskEither()` [doc](https://pkg.go.dev/github.com/samber/mo#NewTaskEither)
- `mo.NewTaskEitherFromIOEither()` [doc](https://pkg.go.dev/github.com/samber/mo#NewTaskEitherFromIOEither)

Methods:

- `.Run()` [doc](https://pkg.go.dev/github.com/samber/mo#TaskEither.Run)
- `.OrElse()` [doc](https://pkg.go.dev/github.com/samber/mo#TaskEither.OrElse)
- `.Match()` [doc](https://pkg.go.dev/github.com/samber/mo#TaskEither.Match)
- `.TryCatch()` [doc](https://pkg.go.dev/github.com/samber/mo#TaskEither.TryCatch)
- `.ToTask()` [doc](https://pkg.go.dev/github.com/samber/mo#TaskEither.ToTask)
- `.ToEither()` [doc](https://pkg.go.dev/github.com/samber/mo#TaskEither.ToEither)

### State[S any, A any]

`State` represents a function `(S) -> (A, S)`, where `S` is state, `A` is result.

Constructors:

- `mo.NewState()` [doc](https://pkg.go.dev/github.com/samber/mo#NewState)
- `mo.ReturnState()` [doc](https://pkg.go.dev/github.com/samber/mo#ReturnState)

Methods:

- `.Run()` [doc](https://pkg.go.dev/github.com/samber/mo#TaskEither.Run)
- `.Get()` [doc](https://pkg.go.dev/github.com/samber/mo#TaskEither.Get)
- `.Modify()` [doc](https://pkg.go.dev/github.com/samber/mo#TaskEither.Modify)
- `.Put()` [doc](https://pkg.go.dev/github.com/samber/mo#TaskEither.Put)

## 🛩 Benchmark

// @TODO

This library does not use `reflect` package. We don't expect overhead.

## 🤝 Contributing

- Ping me on twitter [@samuelberthe](https://twitter.com/samuelberthe) (DMs, mentions, whatever :))
- Fork the [project](https://github.com/samber/mo)
- Fix [open issues](https://github.com/samber/mo/issues) or request new features

Don't hesitate ;)

### With Docker

```bash
docker-compose run --rm dev
```

### Without Docker

```bash
# Install some dev dependencies
make tools

# Run tests
make test
# or
make watch-test
```

## 👤 Contributors

![Contributors](https://contrib.rocks/image?repo=samber/mo)

## 💫 Show your support

Give a ⭐️ if this project helped you!

[![GitHub Sponsors](https://img.shields.io/github/sponsors/samber?style=for-the-badge)](https://github.com/sponsors/samber)

## 📝 License

Copyright © 2022 [Samuel Berthe](https://github.com/samber).

This project is [MIT](./LICENSE) licensed.
