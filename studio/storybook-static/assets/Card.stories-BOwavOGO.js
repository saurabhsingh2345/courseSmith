import{j as e}from"./jsx-runtime-Z5uAzocK.js";import{r as t}from"./index-pP6CS22B.js";import{c as o}from"./cn-BLSKlp9E.js";import{B as C}from"./Button-B8zseNBH.js";import"./_commonjsHelpers-Cpj98o6Y.js";import"./index-7xovqvL3.js";import"./index-EXTQMK5R.js";const n=t.forwardRef(({className:a,...s},r)=>e.jsx("div",{ref:r,className:o("rounded-[var(--radius-lg)] border border-border bg-surface text-fg shadow-sm",a),...s}));n.displayName="Card";const i=t.forwardRef(({className:a,...s},r)=>e.jsx("div",{ref:r,className:o("flex flex-col gap-1.5 p-6",a),...s}));i.displayName="CardHeader";const m=t.forwardRef(({className:a,...s},r)=>e.jsx("h3",{ref:r,className:o("text-lg font-semibold leading-none tracking-tight",a),...s}));m.displayName="CardTitle";const c=t.forwardRef(({className:a,...s},r)=>e.jsx("p",{ref:r,className:o("text-sm text-muted",a),...s}));c.displayName="CardDescription";const l=t.forwardRef(({className:a,...s},r)=>e.jsx("div",{ref:r,className:o("p-6 pt-0",a),...s}));l.displayName="CardContent";const p=t.forwardRef(({className:a,...s},r)=>e.jsx("div",{ref:r,className:o("flex items-center p-6 pt-0",a),...s}));p.displayName="CardFooter";n.__docgenInfo={description:"",methods:[],displayName:"Card"};i.__docgenInfo={description:"",methods:[],displayName:"CardHeader"};m.__docgenInfo={description:"",methods:[],displayName:"CardTitle"};c.__docgenInfo={description:"",methods:[],displayName:"CardDescription"};l.__docgenInfo={description:"",methods:[],displayName:"CardContent"};p.__docgenInfo={description:"",methods:[],displayName:"CardFooter"};const w={title:"Base/Card",component:n,tags:["autodocs"]},d={render:()=>e.jsxs(n,{className:"w-80",children:[e.jsxs(i,{children:[e.jsx(m,{children:"Lesson 3 · Recursion"}),e.jsx(c,{children:"Estimated 12 minutes · 4 exercises"})]}),e.jsx(l,{className:"text-sm text-muted",children:"Break a problem into a base case and a smaller subproblem, then let the call stack do the rest."}),e.jsxs(p,{className:"gap-2",children:[e.jsx(C,{size:"sm",children:"Start"}),e.jsx(C,{size:"sm",variant:"ghost",children:"Preview"})]})]})};var x,f,h;d.parameters={...d.parameters,docs:{...(x=d.parameters)==null?void 0:x.docs,source:{originalSource:`{
  render: () => <Card className="w-80">
      <CardHeader>
        <CardTitle>Lesson 3 · Recursion</CardTitle>
        <CardDescription>Estimated 12 minutes · 4 exercises</CardDescription>
      </CardHeader>
      <CardContent className="text-sm text-muted">
        Break a problem into a base case and a smaller subproblem, then let the call stack do the rest.
      </CardContent>
      <CardFooter className="gap-2">
        <Button size="sm">Start</Button>
        <Button size="sm" variant="ghost">
          Preview
        </Button>
      </CardFooter>
    </Card>
}`,...(h=(f=d.parameters)==null?void 0:f.docs)==null?void 0:h.source}}};const v=["Default"];export{d as Default,v as __namedExportsOrder,w as default};
