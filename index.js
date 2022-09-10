'use strict';

const fs = require('fs');
const path = require('path');
const crypto = require('crypto');

Object.defineProperty(global, 'crypto', {
  value: {
    getRandomValues: (arr) => crypto.randomBytes(arr.length),
  },
});

globalThis.fs = fs

require('./wasm_exec');


module.exports = {
  inspect: (f, m = null) => {
    const go = new Go();
    const exitListener = code => {
      if (code === 0 && !go.exited) {
        go._pendingEvent = { id: 0 };
        go._resume();
      }
    }
    process.on('exit', exitListener);
    // main.that will assign functions to this insance
    const that = {}
    // pass the reference so it's accessible in main.that
    go.that = that

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

        const p = m == null ? that.inspect(f) : that.inspect(f, m);
        p.then(val => {
          resolve(JSON.parse(val));
        })
          .catch(reject)
      })
      .catch(reject)
      .finally(() => {
        that.finish()
        process.removeListener('exit', exitListener)
      });
    });
  }
}
