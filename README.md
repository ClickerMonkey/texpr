# texpr
Text based expression evaluator with types.

### Features
- Types & values (simple fields, methods, or operations) are entirely user defined.
- Easy to understand left to right evaluation & parsing. ex: user.createDate.hour
- Types have parameterized values (methods). ex: today.addDays(2)
- Type methods can be symbols which appear operation like: today.minute+(today.hour*(60))>(120)
- Expressions are case insensitive. ex: TODAY=(today)

```go


```