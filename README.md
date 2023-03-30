# Mail Go

This package is created for supporting my work, I would be glad if the package can help you too.


## Installation
```sh
go get github.com/oinpentuls/mail-go
```

## Usage

Config the smtp server
```go
options := mail.MailOptions{
    Host:     "host",
    Port:     "port",
    Username: "username",
    Password: "password",
}
```

construct new mail
```go
mail := mail.New(options)
mail.SetFrom("hello@hello.com")
mail.SetTo([]string{"hallo@hallo.com"})
mail.SetSubject("Hello")

mail.SetBodyPlainText([]byte("Hello World"))

mail.SetBodyHTML([]byte("<h1>Hello World</h1>"))

err := mail.SetAttachment(filepath.Join("my_file.pdf"))
if err != nil {
    log.Fatal(err)
}

err = mail.Send()
if err != nil {
    log.Fatal(err)
}
```