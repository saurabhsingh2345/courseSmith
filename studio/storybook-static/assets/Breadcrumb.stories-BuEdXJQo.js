import{j as e}from"./jsx-runtime-Z5uAzocK.js";import{r as t}from"./index-pP6CS22B.js";import{S as h}from"./index-7xovqvL3.js";import{c as m}from"./cn-BLSKlp9E.js";import"./_commonjsHelpers-Cpj98o6Y.js";const o=t.forwardRef((r,a)=>e.jsx("nav",{ref:a,"aria-label":"breadcrumb",...r}));o.displayName="Breadcrumb";const u=t.forwardRef(({className:r,...a},s)=>e.jsx("ol",{ref:s,className:m("flex flex-wrap items-center gap-1.5 text-sm text-muted",r),...a}));u.displayName="BreadcrumbList";const d=t.forwardRef(({className:r,...a},s)=>e.jsx("li",{ref:s,className:m("inline-flex items-center gap-1.5",r),...a}));d.displayName="BreadcrumbItem";const c=t.forwardRef(({asChild:r,className:a,...s},f)=>{const x=r?h:"a";return e.jsx(x,{ref:f,className:m("transition-colors duration-[var(--motion-fast)] hover:text-fg",a),...s})});c.displayName="BreadcrumbLink";const l=t.forwardRef(({className:r,...a},s)=>e.jsx("span",{ref:s,role:"link","aria-disabled":"true","aria-current":"page",className:m("font-medium text-fg",r),...a}));l.displayName="BreadcrumbPage";function i({className:r,children:a,...s}){return e.jsx("li",{role:"presentation","aria-hidden":"true",className:m("[&>svg]:size-3.5",r),...s,children:a??"/"})}o.__docgenInfo={description:"",methods:[],displayName:"Breadcrumb"};u.__docgenInfo={description:"",methods:[],displayName:"BreadcrumbList"};d.__docgenInfo={description:"",methods:[],displayName:"BreadcrumbItem"};c.__docgenInfo={description:"",methods:[],displayName:"BreadcrumbLink",props:{asChild:{required:!1,tsType:{name:"boolean"},description:""}}};l.__docgenInfo={description:"",methods:[],displayName:"BreadcrumbPage"};i.__docgenInfo={description:"",methods:[],displayName:"BreadcrumbSeparator"};const _={title:"Base/Breadcrumb",component:o,tags:["autodocs"]},n={render:()=>e.jsx(o,{children:e.jsxs(u,{children:[e.jsx(d,{children:e.jsx(c,{href:"#",children:"Courses"})}),e.jsx(i,{}),e.jsx(d,{children:e.jsx(c,{href:"#",children:"Python"})}),e.jsx(i,{}),e.jsx(d,{children:e.jsx(l,{children:"Recursion"})})]})})};var p,b,B;n.parameters={...n.parameters,docs:{...(p=n.parameters)==null?void 0:p.docs,source:{originalSource:`{
  render: () => <Breadcrumb>
      <BreadcrumbList>
        <BreadcrumbItem>
          <BreadcrumbLink href="#">Courses</BreadcrumbLink>
        </BreadcrumbItem>
        <BreadcrumbSeparator />
        <BreadcrumbItem>
          <BreadcrumbLink href="#">Python</BreadcrumbLink>
        </BreadcrumbItem>
        <BreadcrumbSeparator />
        <BreadcrumbItem>
          <BreadcrumbPage>Recursion</BreadcrumbPage>
        </BreadcrumbItem>
      </BreadcrumbList>
    </Breadcrumb>
}`,...(B=(b=n.parameters)==null?void 0:b.docs)==null?void 0:B.source}}};const L=["Default"];export{n as Default,L as __namedExportsOrder,_ as default};
