package main

import (
	"fmt"
)

func main() {
	fmt.Println("start socks5 port: 1220")
	//socks5.New(1220).Listen()
	a := make([]int, 0)
	fmt.Printf("[]int:%v %d\n", a, cap(a))
	a = append(a, 5)
	a = append(a, 1)
	fmt.Printf("[]int:%v %d\n", a, cap(a))

	var b *[]int = &a
	a = append(a, 1)
	fmt.Printf("b[]int:%v %d\n", *b, cap(*b))
	fmt.Printf("[]int:%v %d\n", a, cap(a))

	c := &b
	a = append(a, 1)
	fmt.Printf("c[]int:%v %d\n", **c, cap(**c))
	fmt.Printf("[]int:%v %d\n", a, cap(a))

	d := &c
	a = append(a, 1)
	fmt.Printf("c[]int:%v %d\n", ***d, cap(***d))
	fmt.Printf("[]int:%v %d\n", a, cap(a))

	var arr1 = []int{1, 2, 3}
	var arr2 = []int{4, 5, 6}
	var arr3 = []int{7, 8, 9}
	var s1 = append(append(arr1, arr2...), arr3...)
	fmt.Printf("s1: %v\n", s1)

}
