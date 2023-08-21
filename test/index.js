var EC = require('elliptic').ec;
var ec = new EC('p256');

// Generate keys
var key1 = ec.genKeyPair();
var key2 = ec.genKeyPair();
key1.getPrivate().
var shared1 = key1.derive(key2.getPublic());
var shared2 = key2.derive(key1.getPublic());
ec.keyFromPublic('04050fa1e797789e4c15a7c29b9a18b8c8badb9b697820bbec3e25b5d002f17fe91231d2d34021d90a6f13cb59e450a9c5154644b55e729351d15899fd1f98fbcd');
console.log('Both shared secrets are BN instances');
console.log(shared1.toString(16));
console.log(shared2.toString(16));
console.log(key2.getPublic().encode("hex"))