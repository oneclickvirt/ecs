'use strict'

// GoECS uses the SSH client only. Keeping this small CommonJS entry point
// avoids loading the optional server and native build helpers.
module.exports = { Client: require('./client.js') }
