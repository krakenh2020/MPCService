from Compiler.library import *

def load_sint():
    v = [sint()]
    input_shares(regint(0), *v)

    return v[0]

def load_sint_array(l):
    a = Array(l, sint)
    @for_range(l)
    def f(i):
        a[i] = load_sint()

    return a

# todo negative
def load_sfix():
    v = load_sint()

    fx = sfix(0)
    fx.v = v

    return fx


def load_sfix_array(l):
    a = Array(l, sfix)
    @for_range(l)
    def f(i):
        a[i] = load_sfix()

    return a

def load_sfix_matrix(dim_x, dim_y):
    a = load_sfix_array(dim_x*dim_y)
    X = sfix.Matrix(dim_x, dim_y)

    @for_range(dim_x)
    def f(i):
        @for_range(dim_y)
        def g(j):
            X[i][j] = a[i * dim_y + j]

    return X

def load_sint_matrix(dim_x, dim_y):
    a = load_sint_array(dim_x*dim_y)
    X = []
    for i in range(dim_x):
        X.append(a[dim_y*i: dim_y*(i+1)])

    return X

def output_sint(res):
    o = [res]
    output_shares(regint(0), *o)

def output_sfix(res):
    o = [res.v]
    output_shares(regint(0), *o)

def output_sint_array(res):
    @for_range(len(res))
    def _(j):
        output_sint(res[j])

def output_sfix_array(res):
    o = [x.v for x in res]
    output_shares(regint(0), *o)


def output_sfix_matrix(res):
    @for_range(len(res))
    def _(j):
        output_sfix_array(res[j])


