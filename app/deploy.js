const { Client } = require('node-scp')
Client({
  host: 'homesecure.dev',
  port: 22,
  username: 'app',
  password: '@app',
  // privateKey: fs.readFileSync('./key.pem'),
  // passphrase: 'your key passphrase',
}).then(client => {
  client.uploadDir('./dist/app', '/opt/app')
        .then(response => {
          client.close() // remember to close connection after you finish
        })
        .catch(error => {})
}).catch(e => console.log(e))