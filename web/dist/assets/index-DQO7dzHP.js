import{j as u,A as d,a as m}from"./index-DTTkVZvB.js";import{r}from"./react-core-CLwr57uy.js";import{m as p}from"./tools-Bia2vmZV.js";import{L as o}from"./semi-ui-BHzOf8tD.js";import"./react-components-B0-Ya59j.js";import"./semantic-7HezVohW.js";const b=()=>{const[t,e]=r.useState(""),[n,i]=r.useState(!1),c=async()=>{e(localStorage.getItem("about")||"");const h=await d.get("/api/about"),{success:l,message:E,data:s}=h.data;if(l){let a=s;s.startsWith("https://")||(a=p.parse(s)),e(a),localStorage.setItem("about",a)}else m(E),e("加载关于内容失败...");i(!0)};return r.useEffect(()=>{c().then()},[]),u.jsx(u.Fragment,{children:n&&t===""?u.jsx(u.Fragment,{children:u.jsxs(o,{children:[u.jsx(o.Header,{children:u.jsx("h3",{children:"关于"})}),u.jsxs(o.Content,{children:[u.jsx("p",{children:"可在设置页面设置关于内容，支持 HTML & Markdown"}),"New-API项目仓库地址：",u.jsx("a",{href:"https://github.com/Calcium-Ion/new-api",children:"https://github.com/Calcium-Ion/new-api"}),u.jsx("p",{children:"NewAPI © 2023 CalciumIon | 基于 One API v0.5.4 © 2023 JustSong。"}),u.jsx("p",{children:"本项目根据MIT许可证授权，需在遵守Apache-2.0协议的前提下使用。"})]})]})}):u.jsx(u.Fragment,{children:t.startsWith("https://")?u.jsx("iframe",{src:t,style:{width:"100%",height:"100vh",border:"none"}}):u.jsx("div",{style:{fontSize:"larger"},dangerouslySetInnerHTML:{__html:t}})})})};export{b as default};
