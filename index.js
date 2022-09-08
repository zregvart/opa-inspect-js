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

module.exports = {
  inspect: (filepath, module = null) => {
    // main.that will assign functions to this insance
    const that = {}

    const go = new Go();
    // pass the reference so it's accessible in main.that
    go.that = that
    process.on('exit', code => {
      if (code === 0 && !go.exited) {
        go._pendingEvent = { id: 0 };
        go._resume();
      }
    });

    return new Promise((resolve, reject) => {
      WebAssembly.instantiate(
        fs.readFileSync(path.resolve(__dirname, 'inspect.wasm')),
        go.importObject
      )
      .then(result => {
        go.run(result.instance);
      })
      .then(() => {
        that.read = (path) => fs.readFileSync(path)

        const val = that.inspect(filepath, module);
        if (val.startsWith("ERR:")) {
          reject(val);
        } else {
          resolve(JSON.parse(val));
        }
      })
      .catch(reject)
      .finally(() => that.finish());
    });
  }
}
