const fs = require('fs');
const path = require('path');
const opa = require('../index');

beforeAll(() => {
    require('events').EventEmitter.defaultMaxListeners = 10;
})

afterAll(() => {
    expect(process.listeners('exit')).toHaveLength(0);
})

test('inspects rego files', async () => {
    const rego = fs.readFileSync(path.join(__dirname, 'example.rego')).toString();
    const json = await opa.inspect('example.rego', rego);
    expect(json).toMatchSnapshot();
});

test('inspects rego files second time', async () => {
    const rego = fs.readFileSync(path.join(__dirname, 'example.rego')).toString();
    const json = await opa.inspect('example.rego', rego);
    expect(json).toMatchSnapshot();
});

test('reports error when parsing', async () => {
    await expect(opa.inspect('example.rego', "bogus")).rejects.toEqual("1 error occurred: example.rego:1: rego_parse_error: package expected");
});

test('inspects rego files read from the filesystem', async () => {
    const json = await opa.inspect(path.join(__dirname, 'example.rego'));
    expect(json).toHaveLength(1);
    expect(json[0]).toMatchSnapshot({
        "location": {
            "file": expect.stringMatching(/.*__test__\/example\.rego/),
        }
    });
});

test('inspects multiple rego files', async () => {
    const json = await opa.inspect([path.join(__dirname, 'example.rego'), path.join(__dirname, 'example2.rego')]);
    expect(json).toHaveLength(2);
    expect(json[0]).toMatchSnapshot({
        "location": {
            "file": expect.stringMatching(/.*__test__\/example\.rego/),
        }
    });
    expect(json[1]).toMatchSnapshot({
        "location": {
            "file": expect.stringMatching(/.*__test__\/example2\.rego/),
        }
    });
});

test('inspects vinyl streams', async () => {
    const vfs = require('vinyl-fs');

    const json = await opa.inspect(vfs.src(path.join(__dirname, '*.rego')));

    expect(json).toHaveLength(2);
    expect(json[0]).toMatchSnapshot({
        "location": {
            "file": expect.stringMatching(/.*__test__\/example\.rego/),
        }
    });
    expect(json[1]).toMatchSnapshot({
        "location": {
            "file": expect.stringMatching(/.*__test__\/example2\.rego/),
        }
    });
});

test.concurrent.each([...Array(25).keys()])('functions concurrently (%#)', async () => {
    const rego = fs.readFileSync(path.join(__dirname, 'example.rego')).toString();
    await expect(opa.inspect('example.rego', rego)).resolves.toHaveProperty('[0].annotations');
});
