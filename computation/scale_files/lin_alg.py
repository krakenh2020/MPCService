import math
from Compiler import mpc_math
from Compiler.library import *


def vectorize_sfix(funct, x):
    n = len(x)
    ret = sfix.Array(n)
    @for_range(n)
    def f(i):
        ret[i] = funct(x[i])
        # print_ln('vectorize %s', ret[i].reveal())

    return ret

def random_vector(dim1, r):
    W = sfix.Array(dim1)
    @for_range(dim1)
    def f(i):
        W[i] = sfix.get_random(-r, r)
    return W

def constant_vector(dim1, r):
    v = sfix.Array(dim1)
    @for_range(dim1)
    def f(i):
        v[i] = sfix(r)
    return v

def constant_matrix(dim1, dim2, r):
    mat = sfix.Matrix(dim1, dim2)
    @for_range(dim1)
    def f(i):
        @for_range(dim2)
        def g(j):
            mat[i][j] = sfix(r)
    return mat

def random_matrix(dim1, dim2, r):
    W = sfix.Matrix(dim1, dim2)
    @for_range(dim1)
    def f(i):
        @for_range(dim2)
        def g(j):
            W[i][j] = sfix.get_random(-r, r)

    return W

def eye_matrix(dim1, dim2):
    W = sfix.Matrix(dim1, dim2)
    @for_range(dim1)
    def f(i):
        @for_range(dim2)
        def g(j):
            if_then(i == j)
            W[i][j] = sfix(1)
            else_then()
            W[i][j] = sfix(0)
            end_if()

    return W

def slice_matrix(M, from_coord, size):
    dim1 = size
    dim2 = len(M[0])

    W = sfix.Matrix(dim1, dim2)
    @for_range(dim1)
    def f(i):
        @for_range(dim2)
        def g(j):
            W[i][j] = M[from_coord + i][j]

    return W

def matrix_mul_vec(M, v):
    rows = len(M)
    cols = len(M[0])
    if cols != len(v):
        return error
    w = sfix.Array(rows)
    @for_range(rows)
    def f(i):
        w[i] = sfix(0)
        @for_range(cols)
        def h(k):
            w[i] = w[i] + (M[i][k] * v[k])

    return w

def mul_vec(x, y):
    n = len(x)
    # print_ln('compare mul_vec %s %s', n, len(y))
    if n != len(y):
        return error
    ret = sfix.Array(n)
    @for_range(n)
    def f(i):
        ret[i] = x[i] * y[i]
    return ret

def matrix_mul_elementwise(M1, M2):
    rows = len(M1)
    cols = len(M1[0])
    if len(M2[0]) != len(M1[0]) or len(M2) != len(M1):
        return error
    W = sfix.Matrix(rows, cols)
    @for_range(rows)
    def f(i):
        @for_range(cols)
        def g(j):
            W[i][j] = M1[i][j] * M2[i][j]

    return W

def matrix_add(M1, M2):
    rows = len(M1)
    cols = len(M1[0])
    if len(M2[0]) != len(M1[0]) or len(M2) != len(M1):
        return error
    W = sfix.Matrix(rows, cols)
    @for_range(rows)
    def f(i):
        @for_range(cols)
        def g(j):
            W[i][j] = M1[i][j] + M2[i][j]

    return W

def matrix_div(M1, M2, epsilon = 0):
    rows = len(M1)
    cols = len(M1[0])
    if len(M2[0]) != len(M1[0]) or len(M2) != len(M1):
        return error
    W = sfix.Matrix(rows, cols)
    @for_range(rows)
    def f(i):
        @for_range(cols)
        def g(j):
            W[i][j] = M1[i][j] / (M2[i][j] + epsilon)

    return W

def matrix_mul_scalar(M1, s):
    rows = len(M1)
    cols = len(M1[0])
    W = sfix.Matrix(rows, cols)
    @for_range(rows)
    def f(i):
        @for_range(cols)
        def g(j):
            W[i][j] = M1[i][j] * s

    return W

def matrix_sqrt(M1):
    rows = len(M1)
    cols = len(M1[0])
    W = sfix.Matrix(rows, cols)
    @for_range(rows)
    def f(i):
        @for_range(cols)
        def g(j):
            W[i][j] = mpc_math.sqrt(M1[i][j])
    return W

def matrix_transpose(M):
    rows = len(M)
    cols = len(M[0])
    W = sfix.Matrix(cols, rows)
    @for_range(cols)
    def f(i):
        @for_range(rows)
        def g(j):
            W[i][j] = M[j][i]
    return W

def matrix_mul(M1, M2):
    rows = len(M1)
    cols = len(M2[0])
    mid = len(M2)
    if len(M1[0]) != mid:
        return error
    W = sfix.Matrix(rows, cols)
    @for_range(rows)
    def f(i):
        @for_range(cols)
        def g(j):
            W[i][j] = sfix(0)
            @for_range(mid)
            def h(k):
                W[i][j] = W[i][j] + (M1[i][k] * M2[k][j])

    return W

def vec_to_matrix(v):
    rows = 1
    cols = len(v)
    W = sfix.Matrix(rows, cols)
    @for_range(cols)
    def f(i):
        W[0][i] = v[i]
    return W

def vector_minus(x, y):
    n = len(x)
    if len(y) != n:
        return error
    ret = sfix.Array(n)
    @for_range(n)
    def f(i):
        ret[i] = x[i] - y[i]
    return ret

def vector_add(x, y):
    n = len(x)
    if len(y) != n:
        return error
    ret = sfix.Array(n)
    @for_range(n)
    def f(i):
        ret[i] = x[i] + y[i]
    return ret

def norm_vec(v):
    norm = MemValue(sfix(0))
    @for_range(len(v))
    def g(i):
        norm.write(norm.read() + v[i] * v[i])
    return norm.read()

def vec_mul_scalar(v,s):
    return vectorize_sfix(lambda x: x*s, v)

def vector_print(v):
    rows = len(v)
    print_ln('vector')
    @for_range(rows)
    def f(i):
        print_ln('%s', v[i].reveal())

def matrix_print(v):
    rows = len(v)
    cols = len(v[0])
    print_ln('matrix')
    @for_range(rows)
    def f(i):
        print_ln('row')
        @for_range(cols)
        def g(j):
            print_ln('%s', v[i][j].reveal())

def normalize(X):
    dim1, dim2 = len(X), len(X[0])
    max_vals = sfix.Array(dim2)
    @for_range(dim2)
    def f(i):
        max_vals[i] = sfix(0)
    @for_range(dim1)
    def f(i):
        @for_range(dim2)
        def g(j):
            max_vals[j] = (X[i][j] > max_vals[j]).if_else(X[i][j], max_vals[j])
    @for_range(dim1)
    def f(i):
        @for_range(dim2)
        def g(j):
            X[i][j] = (sfix(0) < max_vals[j]).if_else(X[i][j] / max_vals[j], X[i][j])

    return max_vals

def eliminate(r1, r2, col, target=0):
    fac = (r2[col]-target) / r1[col]
    # vector_print(r2)
    @for_range(len(r2))
    def f(i):
        r2[i] -= fac * r1[i]
    # vector_print(r2)


def gauss(a):
    @for_range(len(a))
    def f(i):
        # TODO
        # if a[i][i] == 0:
        #     for j in range(i+1, len(a)):
        #         if a[i][j] != 0:
        #             a[i], a[j] = a[j], a[i]
        #             break
        #     else:
        #         raise ValueError("Matrix is not invertible")
        @for_range(len(a)-i-1)
        def g(k):
            j = i + 1 + k
            eliminate(a[i], a[j], i)
    @for_range(len(a))
    def f(k):
        i = len(a)-1-k
        @for_range(i)
        def g(k2):
            j = i-1-k2
            eliminate(a[i], a[j], i)
    @for_range(len(a))
    def f(i):
        eliminate(a[i], a[i], i, target=1)
    return a

def matrix_inverse(a):
    rows = len(a)
    cols = len(a[0])

    eye = eye_matrix(rows, cols)
    tmp = matrix_join_cols(a, eye)

    gauss(tmp)
    ret = sfix.Matrix(rows, cols)
    @for_range(rows)
    def f(i):
        @for_range(cols)
        def g(j):
            ret[i][j] = tmp[i][cols + j]
    return ret

def matrix_join_cols(M1, M2):
    rows = len(M1)
    cols1 = len(M1[0])
    cols2 = len(M2[0])
    ret = sfix.Matrix(rows, cols1+cols2)
    @for_range(rows)
    def f(i):
        @for_range(cols1)
        def g(j):
            ret[i][j] = M1[i][j]
        @for_range(cols2)
        def g(j):
            ret[i][j + cols1] = M2[i][j]

    return ret

def matrix_assign_row(M, v, i):
    @for_range(len(M[0]))
    def f(j):
        M[i][j] = v[j]
