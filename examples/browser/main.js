import * as opa from '@zregvart/opa-inspect'
import inspectWASM from '@zregvart/opa-inspect/inspect.wasm?url'
import Prism from 'prismjs'

opa.inspectWASM(inspectWASM)

const content = document.querySelector('#content')
const output = document.querySelector('#output')
const inspect = () => {
    opa.inspect('file.rego', content.value).then(obj => {
        content.setCustomValidity('')
        const json = JSON.stringify(obj, null, 2)
        const html = Prism.highlight(json, Prism.languages.js, 'js');
        output.innerHTML = html
    }).catch(err => {
        output.innerHTML = ``
        content.setCustomValidity(err)
        content.reportValidity()
    })
}

inspect()

content.addEventListener('input', inspect)
