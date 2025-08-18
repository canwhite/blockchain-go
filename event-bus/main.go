package main

import (
	"fmt"
	"log"
	"sync"
	"time"
)

// ---------- EventCtx ----------
type EventCtx struct {
	// RWMutex（读写锁）和 Mutex（互斥锁）的区别如下：
	mu   sync.RWMutex
	data map[string]any
}

//*EventCtx定义指针,然后内部直接返回指针，外部直接用就ok
func NewEventCtx() *EventCtx { 
	//结构体初始化，make会创建空的初始值
	return &EventCtx{data: make(map[string]any)} 
}

/** ctx的set*/
func (c *EventCtx) Set(k string, v any) {
	c.mu.Lock()
	c.data[k] = v
	c.mu.Unlock()
}

/** ctx的get*/
func (c *EventCtx) Get(k string) (any, bool) {
	c.mu.RLock()
	//comma ok
	v, ok := c.data[k]
	c.mu.RUnlock()
	return v, ok
}

// ---------- EventBus ----------
//返回值()包裹，并不意味返回的就是元组，只是说明是返回多个，这里没有display
type Handler func(ctx *EventCtx, data any)

type EventBus struct {
	mu       sync.RWMutex
	handlers map[string][]Handler// map是系统的，value是[]Handlers
	queues   map[string]chan any // value是chan any
}

func NewEventBus() *EventBus {
	return &EventBus{
		handlers: make(map[string][]Handler),//publish的时候遍历
		queues:   make(map[string]chan any), //传递消息
	}
}

// 订阅
func (eb *EventBus) Subscribe(topic string, fn Handler) {
	//它这种把边界情况放在开始解决
	eb.mu.Lock()
	defer eb.mu.Unlock()

	//将对应handle通过key放进对应数组里
	// 是的，append 返回的是一个新切片。
	// 所以返回值要给到它自己才能更新
	eb.handlers[topic] = append(eb.handlers[topic], fn)
	
	// 这里是在订阅某个 topic 的时候，检查该 topic 是否已经有对应的消息队列（channel）。
	// 如果没有，就为该 topic 创建一个带缓冲区的 channel，并启动一个消费者 goroutine 来处理该 topic 的消息。
	if _, ok := eb.queues[topic]; !ok {
		// 64表示为每个topic分配的消息队列（channel）的缓冲区大小，也就是最多可以缓存64条消息。
		// 如果消息发布速度大于消费速度，超过64条未被消费的消息后，新的消息会被丢弃（见Publish方法的default分支）。
		// 如果不加64，就没有缓冲区，不会丢弃消息
		eb.queues[topic] = make(chan any, 64) //初始化了一个消息channel
		go eb.consumer(topic) 
	}
}

// 发布：传递结果
func (eb *EventBus) Publish(topic string, data any) {
	eb.mu.RLock()
	q, ok := eb.queues[topic]
	eb.mu.RUnlock()
	if !ok {
		log.Printf("topic %s has no subscriber", topic)
		return
	}
	// select 是 Go 语言中用于处理多个 channel 操作的控制结构。它的作用类似于 switch，但每个 case 语句都是一个 channel 的收发操作。
	// select 会监听所有 case 中的 channel，只要其中有一个可以进行（如发送或接收不会阻塞），就会执行对应的 case。
	// 如果多个 case 同时满足条件，select 会随机选择一个执行。
	// 如果所有 case 都阻塞，并且有 default 分支，则会执行 default 分支；如果没有 default，则 select 会一直阻塞，直到有 case 可以执行。
	// 这种机制常用于实现超时处理、消息多路复用等并发场景。
	select {
	case q <- data:
	default:
		log.Printf("topic %s queue full, drop msg", topic)
	}
}

// 顺序消费者：内部自动创建 EventCtx
func (eb *EventBus) consumer(topic string) {
	
	//this side，遍历的是消息缓冲区
	for data := range eb.queues[topic] {
		ctx := NewEventCtx() // 每条消息独立 ctx
		eb.mu.RLock()
		// 这一句的作用是拷贝一份 handlers[topic] 切片，避免后续遍历时原切片被其他 goroutine 修改导致并发问题。
		// 前面的 nil 表示以一个空切片为基础，后面的 ... 是展开 handlers[topic] 这个切片的所有元素，追加到空切片里。
		// 这样 handlers 变量就是 handlers[topic] 的一个副本，后续遍历 handlers 就不会受原切片变化的影响。
		handlers := append([]Handler(nil), eb.handlers[topic]...)
		eb.mu.RUnlock()
		//执行
		for _, h := range handlers {
			h(ctx, data)
		}
	}
}

// ---------- DEMO ----------
func main() {
	bus := NewEventBus()

	bus.Subscribe("order", func(ctx *EventCtx, data any) {
		ctx.Set("step", "validate")
		step, _ := ctx.Get("step")
		fmt.Printf("[1] validate %v, ctx.step=%v\n", data, step)
	})
	bus.Subscribe("order", func(ctx *EventCtx, data any) {
		ctx.Set("step", "save")
		step, _ := ctx.Get("step")
		fmt.Printf("[2] save %v, ctx.step=%v\n", data, step)
	})

	bus.Publish("order", "ORD-2024-001") // 无需 ctx
	bus.Publish("order", "ORD-2024-002") // 每发一次，内部自动新建 ctx

	time.Sleep(time.Second) // 让 goroutine 打印完
}