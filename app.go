package main

import "fmt"

type Person struct {
	name string
	age  int
}

func changeName(p *Person, name string) {
	fmt.Printf("changeName p: %p\n", &p)
	p.name = name // значение меняется, но, не там
}

func getPerson() Person {
	var p Person

	defer changeName(&p, "Alice") // срабатывает после return получая копию p и вызывая у него изменение Alice

	p = Person{
		name: "Bob",
		age:  52,
	}
	fmt.Printf("getP p: %p\n", &p)
	return p // тут возвращается копия p
}

func main() {
	p := getPerson()
	fmt.Printf("main p: %p\n", &p)
	fmt.Println(p)
}
