# gomail

## Sample hello world

```go
package main

import (
    "fmt"

    "github.com/tavo-wasd-gh/gomail"
)

func main() {
    client := gomail.Client("smtp.example.com", "587", "p4ssw0rd")

    err := client.Send(
        "you@example.com",
        []string{"recipient@example.com"},
        "My Subject",
        "Hello world!",
    )

    if err != nil {
        fmt.Println("Error sending email:", err)
    } else {
        fmt.Println("Email sent successfully")
    }
}
```

## With attachments

```go
package main

import (
    "fmt"

    "github.com/tavo-wasd-gh/gomail"
)

func main() {
    client := gomail.Client("smtp.example.com", "587", "p4ssw0rd")

    attachments := map[string]*bytes.Buffer{
        "hello.txt": bytes.NewBuffer([]byte("This is a test file.")),
    }

    err := client.Send(
        "you@example.com",
        []string{"recipient@example.com"},
        "My Subject",
        "Hello world!",
        attachments,
    )

    if err != nil {
        fmt.Println("Error sending email:", err)
    } else {
        fmt.Println("Email sent successfully")
    }
}
```
