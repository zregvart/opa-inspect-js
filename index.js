'use strict';

const fs = require('fs');
const path = require('path');
const crypto = require('crypto');

Object.defineProperty(global, 'crypto', {
  value: {
    getRandomValues: (arr) => crypto.randomBytes(arr.length),
  },
});

require('./wasm_exec');

const terminate = code => {
  if (go.exited) {
    return;
  }

  opa.finish();
  if (code === 0 && !go.exited) {
    go._pendingEvent = { id: 0 };
    go._resume();
  }
}

const go = new Go();
process.on('exit', terminate);

module.exports = {
  inspect: (filename, rego = null) => {
    return new Promise((resolve, reject) => {
      WebAssembly.instantiate(
        fs.readFileSync(path.resolve(__dirname, 'inspect.wasm')),
        go.importObject
      )
      .then(result => {
        go.run(result.instance);
      })
      .then(() => {
        const val = opa.inspect(filename, rego);
        if (val.startsWith("ERR:")) {
          reject(val);
        } else {
          resolve(JSON.parse(val));
        }
      })
      .catch(reject)
      .finally(terminate);
    });
  }
}
