import '@testing-library/jest-dom/vitest';
class TestResizeObserver{observe(){}unobserve(){}disconnect(){}}
Object.defineProperty(globalThis,'ResizeObserver',{writable:true,value:TestResizeObserver});
Object.defineProperty(window,'matchMedia',{writable:true,value:(query:string)=>({matches:false,media:query,onchange:null,addListener:()=>{},removeListener:()=>{},addEventListener:()=>{},removeEventListener:()=>{},dispatchEvent:()=>false})});

