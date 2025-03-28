package handlers

const HTMLAllMetrics = `
<!DOCTYPE html>
<html lang="en">
<head>
    <meta charset="UTF-8">
    <meta name="viewport" content="width=device-width, initial-scale=1.0">
    <title>Metrics</title>
</head>
<body>
    <h1>Metrics</h1>
    <table border="1">
        <tr>
            <th>Name</th>
            <th>Value</th>
        </tr>
        {{range .}}
        <tr>
            <td>{{.Name}}</td>
            <td>{{.Value}}</td>
        </tr>
        {{end}}
    </table>
</body>
</html>
`
