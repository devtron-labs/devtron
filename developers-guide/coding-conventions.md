## About

This coding convention guide aims to compile best approaches for writing readable, idiomatic and efficient Go. We also intend to minimize the guesswork of writing readable and effective Go so that newcomers to the language can avoid common mistakes. This layout is always open for suggestions and improvements. 

## Conventions 

### Choose your receivers wisely 

* Receivers can be thought of as just another arguments for the function. So it can be passed in two ways - `pointer` or `value`. 
* `We would almost always want to use pointer receivers` because value receivers make a copy of the type and pass it to the function which eats up unnecessary memory and slows down the application especially when the receiver is a large struct (as in most of our cases). 
* If some object present in the receiver is to be manipulated, do not do this directly instead make a copy and return the value to propagate the changes.
* Pointer receivers are not concurrency safe if the state is being changed (goroutines, channels) so always use mutex wherever needed. 
* NEVER have methods on both value and pointer receivers. ([Why?](https://go.dev/doc/faq#different_method_sets))

### Arrays vs Slices

* Arrays are values whereas Slices are reference types. Also, array have fixed lengths whereas slices can be expanded.
* Arrays can be used when they are to be used locally in a method, not to be passed and we upfront know the size required (or cases where stack is used instead heap overload).
* Slice can be used when the variable is being passed, where the length is dynamic. But it is always preferable to initiate a slice with length (if it's known) to avoid expanding computation.

### Naming Conventions (for better readability, consistency across the system) 

#### Receiver functions

| File Type   | Receiver Name | Example                           |
|-------------|---------------|-----------------------------------|
| Router      | router        | func(router *ExampleRouterImpl)   |
| RestHandler | handler       | func(handler *ExampleHandlerImpl) |
| Service     | impl          | func(impl *ExampleServiceImpl)    |
| Repository  | repo          | func(repo *ExampleRepositoryImpl) |
| Util        | util          | func(util *ExampleUtilImpl)       |

#### Variables 


| Variable Type     | CASE to be used | Example           |
|-------------------|-----------------|-------------------|
| Private Variables | camelCase       | myPrivateVariable |
| Public Variables  | PascalCase      | MyPublicVariable  |

#### Constants

* Use `PascalCase` for naming.
* Use [Iota Constants](https://go.dev/doc/effective_go#constants).
* Always use Constants instead of hard-coding strings.

#### Avoid using [Predeclared Identifiers](https://go.dev/ref/spec#Predeclared_identifiers)

### Errors

* Use our own [custom error](https://github.com/devtron-labs/devtron/blob/main/internal/util/ErrorUtil.go#L25) as much as we can instead of using `fmt.Errorf` or `errors.New` etc.
  * We must define and use our own different codes as much as possible. [Example](https://github.com/devtron-labs/devtron/blob/main/internal/constants/InternalErrorCode.go).
  * InternalMessage must only be used for developer's understanding. Example - `Team creation failed`, here user is unaware with the term `Team` as it is used to denote `Projects` at BE.
  * UserMessage must always be filled with the simplest of words and only with terms known to the user. Example `Project creation failed, project already exists.`
  
* Handle errors and avoid nesting (The reader has a cognitive load to process when using an if-statement, which demands more power to run our code). 

```
Not Preferred - 
    err := request()
    if err != nil {
        // handling error 
    } else {
        // some code
    }
 
Preferred - 
    err := request()
    if err != nil {
        // handling error
        return // or continue 
    } 
    // some code
   
```

## Directory Structure 


### `/api`

Current implementation - 

* This directory is currently used for storing `Routers` and `RestHandlers`. Most of them are stored in the `/router` and `/restHandler` package.

Aim - 

* The router and restHandler must be categorised according to different entities. We aim to create subdirectories for routers and restHandlers according to the entity they belong to.

### `/internal`

* This directory enable us to export code for reuse in our project while reducing our public API.

Current implementation -

* Third party library code. Example - GitOps services, Argo util etc.
* DB Repositories (in the subdirectory `/sql`).

Aim - 

* Phase out db repositories from this location and keep them along with their respective services.
* To keep only private applications and library code in this directory. Go itself enforces keeping of private code by - 

```
When the go command sees an import of a package with internal in its path,
it verifies that the package doing the import is within the tree rooted at the parent of the internal directory. 
For example, a package .../a/b/c/internal/d/e/f can be imported only by code in the directory tree rooted at .../a/b/c.
It cannot be imported by code in .../a/b/g or in any other repository.
```

### `/pkg`

* Library code that's ok to use by external applications is placed in this directory. Other projects will import these libraries expecting them to work. This directory is a good way to explicitly communicate that the code in that directory is safe for use by others.

Current implementation - 
*  Services and Repositories are placed in this directory. 

Aim - 

* We need to continue the current structure along with generalising the use of this directory to follow the first most point of this directory's description.
* Objects used in the code must be preferably kept in `bean.go`.

### `/vendor` (non-editable)

Application dependencies (managed by [`Go Modules`](https://github.com/golang/go/wiki/Modules) dependency management). The `go mod vendor` command will create the `/vendor` directory for you if not present. 