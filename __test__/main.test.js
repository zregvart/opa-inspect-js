import * as fs from 'node:fs';
import * as events from 'node:events';
import * as url from 'node:url';
import * as vfs from 'vinyl-fs';
import Vinyl from 'vinyl';
import * as opa from '../main';

testFile = (f) => url.fileURLToPath(new url.URL(f, import.meta.url))

describe("nodejs", () => {
    beforeAll(() => {
        events.EventEmitter.defaultMaxListeners = 10;
    })

    afterAll(() => {
        expect(process.listeners('exit')).toHaveLength(0);
    })

    test('inspects rego files', async () => {
        const rego = fs.readFileSync(testFile('./example.rego')).toString();
        const json = await opa.inspect('example.rego', rego);
        expect(json).toMatchSnapshot();
    });

    test('inspects rego files second time', async () => {
        const rego = fs.readFileSync(testFile('./example.rego')).toString();
        const json = await opa.inspect('example.rego', rego);
        expect(json).toMatchSnapshot();
    });

    test('reports error when parsing', async () => {
        await expect(opa.inspect('example.rego', "bogus")).rejects.toEqual("1 error occurred: example.rego:1: rego_parse_error: package expected");
    });

    test('inspects rego files read from the filesystem', async () => {
        const json = await opa.inspect(testFile('./example.rego'));
        expect(json).toHaveLength(1);
        expect(json[0]).toMatchSnapshot({
            "location": {
                "file": expect.stringMatching(/\/example\.rego$/),
            }
        });
    });

    test('inspects multiple rego files', async () => {
        const json = await opa.inspect([testFile('./example.rego'), testFile('./example2.rego')]);
        expect(json).toHaveLength(2);
        expect(json[0]).toMatchSnapshot({
            "location": {
                "file": expect.stringMatching(/\/example\.rego$/),
            }
        });
        expect(json[1]).toMatchSnapshot({
            "location": {
                "file": expect.stringMatching(/\/example2\.rego$/),
            }
        });
    });

    test('inspects vinyl streams', async () => {
        const json = await opa.inspect(vfs.src('__test__/*.rego'));

        expect(json).toHaveLength(2);
        expect(json[0]).toMatchSnapshot({
            "location": {
                "file": expect.stringMatching(/\/example\.rego$/),
            }
        });
        expect(json[1]).toMatchSnapshot({
            "location": {
                "file": expect.stringMatching(/\/example2\.rego$/),
            }
        });
    });

    test('multiple vinyl files', async () => {
        const base = url.fileURLToPath(import.meta.url)
        const example = new Vinyl({
            cwd: '/',
            base: base,
            path: `/${base}/example.rego`,
            contents: fs.readFileSync(testFile('./example.rego', import.meta.url))
        });

        const example2 = new Vinyl({
            cwd: '/',
            base: base,
            path: `/${base}/example2.rego`,
            contents: fs.readFileSync(testFile('./example2.rego'))
        });

        const json = await opa.inspect([example, example2]);

        expect(json).toHaveLength(2);
        expect(json[0]).toMatchSnapshot({
            "location": {
                "file": expect.stringMatching(/\/example\.rego$/),
            }
        });
        expect(json[1]).toMatchSnapshot({
            "location": {
                "file": expect.stringMatching(/\/example2\.rego$/),
            }
        });
    });

    test.concurrent.each([...Array(25).keys()])('functions concurrently (%#)', async () => {
        const rego = fs.readFileSync(new url.URL('./example.rego', import.meta.url)).toString();
        await expect(opa.inspect('example.rego', rego)).resolves.toHaveProperty('[0].annotations');
    });

    test('package metadata', async () => {
        const json = await opa.inspect('example.rego', `#
# METADATA
# title: title
# description: description
package example`);
        expect(json).toMatchSnapshot();
    })
});
