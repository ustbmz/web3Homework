package main

import (
	"fmt"
	"math"
)

// Shape 形状接口：只要实现 Area 和 Perimeter，就是一种 Shape。
// Go 没有 class / implements 关键字，接口满足是隐式的（鸭子类型）。
type Shape interface {
	Area() float64
	Perimeter() float64
}

// Rectangle 矩形：用长、宽描述。
type Rectangle struct {
	Width  float64
	Height float64
}

// Area 矩形面积 = 长 × 宽。
// 值接收者：方法不修改结构体，调用方用值或指针都能调用。
func (r Rectangle) Area() float64 {
	return r.Width * r.Height
}

// Perimeter 矩形周长 = 2 × (长 + 宽)。
func (r Rectangle) Perimeter() float64 {
	return 2 * (r.Width + r.Height)
}

// Circle 圆：用半径描述。
type Circle struct {
	Radius float64
}

// Area 圆面积 = πr²。
func (c Circle) Area() float64 {
	return math.Pi * c.Radius * c.Radius
}

// Perimeter 圆周长 = 2πr。
func (c Circle) Perimeter() float64 {
	return 2 * math.Pi * c.Radius
}

// printShape 面向接口编程：调用方不关心具体是矩形还是圆。
func printShape(name string, s Shape) {
	fmt.Printf("%s  面积 = %.2f  周长 = %.2f\n", name, s.Area(), s.Perimeter())
}

// Person 人：被 Employee 组合（嵌入）进去。
type Person struct {
	Name string
	Age  int
}

// Employee 员工：匿名嵌入 Person，因此可以直接写 e.Name、e.Age。
// 这是「组合」而不是继承：Employee「有一个」Person，不是「是一个」Person。
type Employee struct {
	Person     // 匿名嵌入，字段和方法都会提升到外层
	EmployeeID string
}

// PrintInfo 值接收者：只读打印，不需要改 Employee。
func (e Employee) PrintInfo() {
	fmt.Printf("员工信息: ID=%s  姓名=%s  年龄=%d\n", e.EmployeeID, e.Name, e.Age)
}

func main() {
	// ---------- 题目 1：Shape 接口 + Rectangle / Circle ----------
	fmt.Println("===== Shape 接口 =====")
	rect := Rectangle{Width: 4, Height: 3}
	circle := Circle{Radius: 2}
	printShape("矩形 4×3", rect)
	printShape("圆 r=2", circle)

	// ---------- 题目 2：组合 Person 得到 Employee ----------
	fmt.Println("\n===== 组合 Employee =====")
	emp := Employee{
		Person:     Person{Name: "张三", Age: 28},
		EmployeeID: "E1001",
	}
	emp.PrintInfo()
}
