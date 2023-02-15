import * as go from './wasm_exec.js';

const inspect = (f, m) => {
    const goruntime = new Go();
    // main.that will assign functions to this insance
    const that = {}
    // pass the reference so it's accessible in main.that
    goruntime.that = that

    return new Promise((res, rej) => {
        WebAssembly.instantiateStreaming(
            fetch('./inspect.wasm'),
            goruntime.importObject
        )
        .then(result => {
            goruntime.run(result.instance);
        })
        .then(() => {
            const p = that.inspect(f, m);
            p.then(val => {
                res(JSON.parse(val));
            })
            .catch(rej)
        })
        .catch(rej)
        .finally(() => {
            that.finish && that.finish()
        });
    });
}

export { inspect }
