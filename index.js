"use strict";

const fs = require("fs");
const path = require("path");
const crypto = require("crypto");

Object.defineProperty(global, "crypto", {
  value: {
    getRandomValues: (arr) => crypto.randomBytes(arr.length),
  },
});

require("./wasm_exec");

module.exports = (cb) => {
  (async function () {
    const go = new Go();
    go.exit = process.exit;
    process.on("exit", (code) => {
      if (code === 0 && !go.exited) {
        go._pendingEvent = { id: 0 };
        go._resume();
      }
    });

    let result = await WebAssembly.instantiate(
      fs.readFileSync(path.resolve(__dirname, "inspect.wasm")),
      go.importObject
    );
    go.run(result.instance);
  })().then(() => {
    cb(opa.inspect);
    opa.finish();
  });
};
