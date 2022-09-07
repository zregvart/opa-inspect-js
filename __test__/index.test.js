const fs = require('fs');
const path = require('path');
const opa = require('../index');

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
    await expect(opa.inspect('example.rego', "bogus")).rejects.toEqual("ERR: 1 error occurred: example.rego:1: rego_parse_error: package expected");
});

test('inspects rego files read from the filesystem', async () => {
    const json = await opa.inspect(path.join(__dirname, 'example.rego'));
    expect(json).toMatchSnapshot();
});
