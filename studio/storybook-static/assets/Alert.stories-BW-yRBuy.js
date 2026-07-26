import{j as r}from"./jsx-runtime-Z5uAzocK.js";import{r as c}from"./index-pP6CS22B.js";import{c as A}from"./index-EXTQMK5R.js";import{c as d}from"./cn-BLSKlp9E.js";import{c as w}from"./createLucideIcon-DfQGSEAs.js";import"./_commonjsHelpers-Cpj98o6Y.js";/**
 * @license lucide-react v0.454.0 - ISC
 *
 * This source code is licensed under the ISC license.
 * See the LICENSE file in the root directory of this source tree.
 */const b=w("Info",[["circle",{cx:"12",cy:"12",r:"10",key:"1mglay"}],["path",{d:"M12 16v-4",key:"1dtifu"}],["path",{d:"M12 8h.01",key:"e9boi3"}]]),h=A("relative w-full rounded-[var(--radius-md)] border p-4 [&>svg]:absolute [&>svg]:left-4 [&>svg]:top-4 [&>svg]:size-4 [&>svg~*]:pl-7",{variants:{variant:{default:"border-border bg-surface text-fg",info:"border-info/40 bg-info/10 text-info",success:"border-success/40 bg-success/10 text-success",warning:"border-warning/40 bg-warning/10 text-warning",error:"border-error/40 bg-error/10 text-error"}},defaultVariants:{variant:"default"}}),t=c.forwardRef(({className:e,variant:a,...s},v)=>r.jsx("div",{ref:v,role:"alert",className:d(h({variant:a}),e),...s}));t.displayName="Alert";const n=c.forwardRef(({className:e,...a},s)=>r.jsx("h5",{ref:s,className:d("mb-1 font-medium leading-none tracking-tight",e),...a}));n.displayName="AlertTitle";const l=c.forwardRef(({className:e,...a},s)=>r.jsx("div",{ref:s,className:d("text-sm opacity-90 [&_p]:leading-relaxed",e),...a}));l.displayName="AlertDescription";t.__docgenInfo={description:"",methods:[],displayName:"Alert",composes:["VariantProps"]};n.__docgenInfo={description:"",methods:[],displayName:"AlertTitle"};l.__docgenInfo={description:"",methods:[],displayName:"AlertDescription"};const I={title:"Base/Alert",component:t,tags:["autodocs"],argTypes:{variant:{options:["default","info","success","warning","error"],control:{type:"select"}}}},i={args:{variant:"info"},render:e=>r.jsxs(t,{...e,className:"max-w-md",children:[r.jsx(b,{}),r.jsx(n,{children:"Render queued"}),r.jsx(l,{children:"Your course video will be ready in a few minutes."})]})},o={render:()=>r.jsx("div",{className:"grid max-w-md gap-3",children:["default","info","success","warning","error"].map(e=>r.jsxs(t,{variant:e,children:[r.jsx(n,{className:"capitalize",children:e}),r.jsxs(l,{children:["This is a ",e," alert."]})]},e))})};var m,p,f;i.parameters={...i.parameters,docs:{...(m=i.parameters)==null?void 0:m.docs,source:{originalSource:`{
  args: {
    variant: 'info'
  },
  render: args => <Alert {...args} className="max-w-md">
      <Info />
      <AlertTitle>Render queued</AlertTitle>
      <AlertDescription>Your course video will be ready in a few minutes.</AlertDescription>
    </Alert>
}`,...(f=(p=i.parameters)==null?void 0:p.docs)==null?void 0:f.source}}};var u,g,x;o.parameters={...o.parameters,docs:{...(u=o.parameters)==null?void 0:u.docs,source:{originalSource:`{
  render: () => <div className="grid max-w-md gap-3">
      {(['default', 'info', 'success', 'warning', 'error'] as const).map(v => <Alert key={v} variant={v}>
          <AlertTitle className="capitalize">{v}</AlertTitle>
          <AlertDescription>This is a {v} alert.</AlertDescription>
        </Alert>)}
    </div>
}`,...(x=(g=o.parameters)==null?void 0:g.docs)==null?void 0:x.source}}};const R=["Default","AllVariants"];export{o as AllVariants,i as Default,R as __namedExportsOrder,I as default};
