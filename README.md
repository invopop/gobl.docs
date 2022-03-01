# Overview

### Objective

The objective that underpins GOBL, which we pronounce as "gobble", is to simplify the creation and sharing of electronic business documents so that they are easily accessible using modern standards and development techniques.

Our focus initially is on invoicing, as we believe this is the area of electronic business documents that needs the most attention. The aim however is to create a library that is flexible enough to represent any business need.&#x20;

### History

A lot has been done in recent decades to make electronic invoicing and other business documents more common place, yet the harsh reality is that companies are simply not making the switch, unless forced to. The reason for this in our opinion, is simply because of a lack of global standards that are easy to use by the people who build solutions: developers.

The potential benefits of wide-spread electronic formats is massive. Imagine how much time could be saved dealing with expenses if all the receipts or invoices we receive contained structured data. PDFs are more convenient, but the data they contain cannot be easily extracted, so business owners and accountants spend hours manually extracting data to be inserted manually into their accounting platforms.

The GOBL team however doesn't just aspire to make business easier, we also see huge potential benefits to individuals and personal finances. Electronic payments are ubiquitous in large parts of the world, yet we still for the most part are given receipts on paper. Paper is hopelessly inefficient for being able to extract data, not to mention the environmental factors of waste. We envision a future whereby everything we buy generates an electronic receipt that can be sent to personal accounting platforms of our choosing.

Conventional standards focus on defining a schema, usually based on XML, and leave the implementation and localisation aspects to other libraries and local state agencies. Given the flexibility of name-spacing and regional extensions, the results are usually difficult and time consuming to use and implement. They also lack tax definitions and validation rules, so it's up to developers to ensure what they're creating will actually work, usually through trial and error.

### Future

For GOBL we wanted to leverage Open Source practices to create a library that defines and serves as the base implementation of the standard itself, in code. Contributing is thus as simple as sending a pull request for the GOBL maintenance team to review and accept.

In most standards, localisation is excluded from the definitions. It makes sense; changes in taxes and rules are hard to keep up with using traditional schemas.

However, we see huge time saving benefits from the Open Source approach by allowing local customisations to be defined in the core library itself, so they can be verified and kept up to date by anyone. With GOBL, there is no need to ask accountants or legal departments for the latest tax definitions nor updates, it's all in code.

Trust is a key element of any electronic format. For us it was very important to make it trivial to digitally sign contents. GOBL leverages JSON Web Signatures and ECDSA keys by default.

In summary, GOBL is our proposal for the future of business documents.
