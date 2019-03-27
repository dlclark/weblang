package compiler

import "testing"

func TestHelloWorldConst(t *testing.T) {
	c := &Compiler{}
	res, err := c.CompilePage(`<!doctype html>
<html>
    <head>
        <meta charset="utf-8">
        <title>Hello Weblang</title>
        <link rel="stylesheet" href="index.css">
    </head>
    <body>
        <h1>{{message}}</h1>
        Our first template!
    </body>
</html>`, `const message = "Hello World!"`)

	if err != nil {
		t.Fatalf("compile error: %v", err)
	}

	if res == nil {
		t.Fatal("no error, but no results")
	}

	if res.Js != "" {
		t.Fatalf("Incorrect JS, got '%v'", res.Js)
	}

	if res.HTML != `<!doctype html>
    <html>
        <head>
            <meta charset="utf-8">
            <title>Hello Weblang</title>
            <link rel="stylesheet" href="index.css">
        </head>
        <body>
            <h1>Hello World!</h1>
            Our first template!
        </body>
    </html>` {
		t.Fatalf("Incorrect HTML, got '%v'", res.HTML)
	}
}

func TestHelloWorldVar(t *testing.T) {
	c := &Compiler{}
	res, err := c.CompilePage(`<!doctype html>
<html>
    <head>
        <meta charset="utf-8">
        <title>Hello Weblang</title>
        <link rel="stylesheet" href="index.css">
    </head>
    <body>
        <h1>{{message}}</h1>
        Our first template!
    </body>
</html>`, `var message = "Hello World!"`)

	if err != nil {
		t.Fatalf("compile error: %v", err)
	}

	if res == nil {
		t.Fatal("no error, but no results")
	}

	if res.Js != "" {
		t.Fatalf("Incorrect JS, got '%v'", res.Js)
	}

	if res.HTML != `<!doctype html>
    <html>
        <head>
            <meta charset="utf-8">
            <title>Hello Weblang</title>
            <link rel="stylesheet" href="index.css">
        </head>
        <body>
            <h1></h1>
            Our first template!
        </body>
    </html>` {
		t.Fatalf("Incorrect HTML, got '%v'", res.HTML)
	}
}
