import * as http from 'node:http';
import * as fs from 'node:fs';
import * as path from 'node:path';
import * as puppeteer from 'puppeteer';
import * as url from 'node:url';

const testHTML = `<!doctype html>
<title>Test</title>
<h1>Test page</h1>
`

const exampleRego = `package example

# METADATA
# title: Task bundle was not used or is not defined
# description: |-
#   Check for existence of a task bundle. Enforcing this rule will
#   fail the contract if the task is not called from a bundle.
# custom:
#   short_name: disallowed_task_reference
#   failure_msg: Task '%s' does not contain a bundle reference
#
deny[msg] {
    msg := "nope"
}`

describe("browser", () => {
    let browser, server;

    beforeAll(async () => {
        server = http.createServer((req, res) => {
            if (req.url === '/') {
                res.writeHead(200, { 'Content-Type': 'text/html' })
                res.end(testHTML, 'utf-8')
                return
            }

            let file = path.join(path.dirname(url.fileURLToPath(import.meta.url)), '..' , req.url);

            const contentType = {
                '.js': 'application/javascript',
                '.wasm': 'application/wasm',
            }[path.extname(file)];

            fs.readFile(file, function (err, content) {
                if (err && err.code == 'ENOENT') {
                    console.log(`Not found ${file}: ${err}`);
                    res.writeHead(200, { 'Content-Type': contentType });
                    res.end();
                } else {
                    res.writeHead(200, { 'Content-Type': contentType });
                    res.end(content, 'utf-8');
                }
            });

        }).listen(8125);

        browser = await puppeteer.launch();
    })

    afterAll(async () => {
        await server.close();
        await browser.close();
    })

    test('inspects rego', async () => {
        const page = await browser.newPage();
        page.on('pageerror', err => {
            throw new Error(err);
        });
        page.on('error', err => {
            throw new Error(err);
        });
        page.on('requestfailed', err => {
            throw new Error(err);
        });
        await page.goto('http://localhost:8125/')
        page.on('console', (c) => console.log('browser console: ' + c.text()));

        await page.addScriptTag({
            content: `
            import * as opa from '../browser.js';

            globalThis.opa = opa;
            `,
            type: 'module'
        });

        await page.waitForNetworkIdle();

        const json = await page.evaluate((exampleRego) => opa.inspect('example.rego', exampleRego), exampleRego);
        expect(json).toMatchSnapshot();
    });
});
