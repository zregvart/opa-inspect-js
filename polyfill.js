import * as fs from 'node:fs';
import * as crypto from 'node:crypto';

Object.defineProperty(globalThis, 'crypto', {
    value: {
        getRandomValues: (arr) => crypto.randomBytes(arr.length),
    },
});

globalThis.fs = fs;
