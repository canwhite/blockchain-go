package main

import (
	"context"
	"fmt"
	"os"
	"io"
	"sync"
	"bytes"
	"json"
)

//===========1.优先使用值类型，只在必要的时候使用指针=========================
//===== 1) 默认使用值类型，只在必要时使用指针,只在必要的时候使用指针
type Point struct{
	X,Y int
}

//值接收者
func (p Point) Move(dx,dy int) Point{
	return Point{p.X + dx , p.Y +dy}
}

//===== 2) 在需要写的时候，我们要用指针
type Counter struct {
	value int
}

//指针接收者
func (c *Counter) Increment(){
	c.value ++ 
}

func (c *Counter) Value() int{
	return c.value
}
// ===== 3）一种好的实践：接口返回时考虑是否需要指针，也就是需要读，还是需要写
type User struct{
	//定义的时候可以省逗号，只有实例化的时候特殊，逗号一个都不能省
	ID int
	Name string  
}
func GetUserPtr(id int)(*User,error){
	return &User{	
		ID:id,
		Name:"张三",
	},nil
}
func GetUserValue(id int) (User, error) {
    // 从数据库获取...
    return User{ID: id, Name: "张三",}, nil
}

//===========2. 管理好资源的声明周期，defer 提前释放，提前判断边界err=============
func ReadFile(fileName string)(string ,error){
	f,err := os.Open(fileName)
	//提前判断和提前释放
	if err != nil {
		return "",err
	}
	defer f.Close()

	data,err := io.ReadAll(f)
	if err != nil{
		return "",err
	}	
	return string(data),nil
}

//===========3. 管理好goroutine的声明周期，使用context=====================
//注意：分清楚chan是出的还是入的
func Worker(ctx context.Context,ch <-chan int){
	for {
		select{
		case <- ctx.Done():
			return;
		case n := <-ch:
			fmt.Println("处理:", n)
		}
	}	
}

func UseWorker(){
	ctx, cancel := context.WithCancel(context.Background());
	defer cancel()
	ch := make(chan int)
	go Worker(ctx,ch)
	
	for i := 0;i < 5; i++{
		ch <- i 
	}
}


//===========4. 对于长期存在的数据进行大小限制，主要是防止内存泄漏和内存溢出问题==
type Cache struct{
	mu sync.Mutex
	items map[string]interface{}
	maxSize int //做大小限制
}

func NewCache(maxSize int) *Cache{
	return &Cache{
		items:make(map[string]interface{}),
		maxSize:maxSize,//这里的逗号不能省略，编译器限制
	}
}

func (c *Cache)Set(key string,value interface{}){
	c.mu.Lock()
	defer c.mu.Unlock()
	//判断maxSize 	
	if(len(c.items) >= c.maxSize){
		// 这里其实并不是“随机删除”，而是因为Go语言的map在遍历时，key的顺序是随机的（每次遍历顺序都可能不同），
		// 所以用 for k := range c.items 取到的第一个key并没有确定的顺序。
		// 这样删除的元素在逻辑上是“随机”的，但实际上是取到的第一个遍历到的key。
		// 如果需要真正的LRU等淘汰策略，需要额外的数据结构来维护顺序。
		for k := range c.items{
			delete(c.items,k)
			break
		}
	}
	//set
	c.items[key] = value
}

func (c *Cache)Get(key string) interface{}{
	c.mu.Lock()
	defer c.mu.Unlock()
	//记一下map有comma ok，所以go语言的精髓的一部分就是判断在前

	value, ok := c.items[key]
	if !ok {
		return nil
	}
	//返回interface{},意味着返回任意类型
	return value
}

//===========5.对于频繁创建和销毁的对象，建议使用对象池，减少gc压力，提高性能=====
// sync.Pool 提供了一个轻量级的对象池，允许我们将用完的对象放回池中，下次需要时直接复用，避免了重复分配和回收。
var bufferPool = sync.Pool{
	New: func() interface{}{
		// buffer（缓冲区）是一个用于临时存储数据的内存区域。可以高效地拼接、追加、读取字节数据，
		// 常用于 I/O 操作、字符串拼接、序列化等场景，避免频繁的内存分配和拷贝。
		return new(bytes.Buffer)
	},
}

func getBuffer() *bytes.Buffer {
	//.(*bytes.Buffer)是断言，因为Get得到的是interface，它可以是任何类型
	return bufferPool.Get().(*bytes.Buffer)
}

func putBuffer(buf *bytes.Buffer){
	buf.Reset()
    bufferPool.Put(buf)
}

func processData(data []byte){
	buf := getBuffer()
    defer putBuffer(buf)

    buf.Write(data)
    // 处理数据...
}


// bufferPool 除了可以反复声明 bytes.Buffer 实例用于高效复用（如序列化、临时缓冲等），还可以用来存放和复用其他类型的对象，减少频繁分配和回收带来的GC压力。
type MyStruct struct {
	Data []int
}

var myStructPool = sync.Pool{
	New: func() interface{} {
		return &MyStruct{
			Data: make([]int, 0, 100),
		}
	},
}

// 使用示例
func UseMyStructPool() {
	obj := myStructPool.Get().(*MyStruct)
	// 使用 obj
	obj.Data = obj.Data[:0] // 清空数据，准备复用
	// ... 业务逻辑 ...
	myStructPool.Put(obj) // 用完放回池中
	
}

func main(){
	p := Point{1,2}
	p = p.Move(3,4)
	fmt.Println(p)

	c := &Counter{}
	c.Increment()
	fmt.Println(c.Value()) // 输出: 1

	UseWorker()


	// 这里演示Cache的使用
	cache := NewCache(2)
	cache.Set("a", 100)
	cache.Set("b", 200)
	fmt.Println("a:", cache.Get("a")) // 输出: a: 100
	fmt.Println("b:", cache.Get("b")) // 输出: b: 200

	// 添加第三个元素，会触发淘汰策略
	cache.Set("c", 300)
	fmt.Println("a:", cache.Get("a")) // 可能为nil（被淘汰）
	fmt.Println("b:", cache.Get("b")) // 可能还在
	fmt.Println("c:", cache.Get("c")) // 输出: c: 300

}