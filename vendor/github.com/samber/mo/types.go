package mo

type f0[R any] func() R
type f1[R any, A any] func(A) R
type f2[R any, A any, B any] func(A, B) R
type f3[R any, A any, B any, C any] func(A, B, C) R
type f4[R any, A any, B any, C any, D any] func(A, B, C, D) R
type f5[R any, A any, B any, C any, D any, E any] func(A, B, C, D, E) R

type ff0[R any] func() *Future[R]
type ff1[R any, A any] func(A) *Future[R]
type ff2[R any, A any, B any] func(A, B) *Future[R]
type ff3[R any, A any, B any, C any] func(A, B, C) *Future[R]
type ff4[R any, A any, B any, C any, D any] func(A, B, C, D) *Future[R]
type ff5[R any, A any, B any, C any, D any, E any] func(A, B, C, D, E) *Future[R]

type fe0[R any] func() (R, error)
type fe1[R any, A any] func(A) (R, error)
type fe2[R any, A any, B any] func(A, B) (R, error)
type fe3[R any, A any, B any, C any] func(A, B, C) (R, error)
type fe4[R any, A any, B any, C any, D any] func(A, B, C, D) (R, error)
type fe5[R any, A any, B any, C any, D any, E any] func(A, B, C, D, E) (R, error)
