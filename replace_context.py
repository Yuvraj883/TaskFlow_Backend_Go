import sys

files = [
    "internal/handlers/project.go",
    "internal/handlers/task.go",
]

for f in files:
    with open(f, "r") as file:
        content = file.read()
    
    content = content.replace("context.Background()", "c.Request.Context()")
    
    with open(f, "w") as file:
        file.write(content)
