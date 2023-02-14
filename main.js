import * as fs from 'node:fs';
import * as polyfill from './polyfill.js';
import * as go from './wasm_exec.js';

const inspect = (f, m = null) => {
  const goruntime = new Go();
  const exitListener = code => {
    if (code === 0 && !goruntime.exited) {
      goruntime._pendingEvent = { id: 0 };
      goruntime._resume();
    }
  }
  process.on('exit', exitListener);
  // main.that will assign functions to this insance
  const that = {}
  // pass the reference so it's accessible in main.that
  goruntime.that = that

  return new Promise((res, rej) => {
    WebAssembly.instantiate(
      fs.readFileSync(new URL('./inspect.wasm', import.meta.url)),
      goruntime.importObject
    )
    .then(result => {
      goruntime.run(result.instance);
    })
    .then(() => {
      that.read = (path) => fs.readFileSync(path)

      const p = m == null ? that.inspect(f) : that.inspect(f, m);
      p.then(val => {
        res(JSON.parse(val));
      })
      .catch(rej)
    })
    .catch(rej)
    .finally(() => {
      that.finish && that.finish()
      process.removeListener('exit', exitListener)
    });
  });
}

export { inspect }
