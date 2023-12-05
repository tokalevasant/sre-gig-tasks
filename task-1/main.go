package main

import "fmt"

func main(){
    task_1_1()
    fmt.Println("")
    task_1_2()
}

func task_1_2() {
	var Sports [5] string
    Sports[0] = "Cricket"
    Sports[1] = "Soccer"
    Sports[2] = "Tennis"
    Sports[3] = "Hockey"
    Sports[4] = "Basketball"

    for i,v := range Sports {
        fmt.Printf("This is %#v and it's index is %d\n",v,i)
    }
}

func task_1_1() {
	var Menu [2] string
    Menu[0] = "hamburger"
    Menu[1] = "salad"

    for _,v := range Menu {
        fmt.Printf("Food: %v\n",v)
    }
}