# vpcdelorean

## Background

AWS re:Invent 2022 was pretty cool. There were a bunch of neat releases. In the 
network space, [VPC Lattice][vpc-lattice] looks very useful. Disappointingly, 
Amazon played it conservative this year and none of their new services break
causality or travel even a _teeny_ distance back in time. That's why I've built
VPC DeLorean: a throwback to the cloud's heyday when physics were suggestions,
not "laws".

## What does it do?

`vpcdelorean` accelerates your packets to 88 miles per hour, so instead of this
snooze fest:

![before](/docs/before.png)

you get responses to your ping packets before you even send them:

![after](/docs/after.png)

Annoyingly, Linux expects a mostly linear progression of time and takes 
countermeasures when time travel is detected. I'm not sure what those countermeasures
_are_, but they sound cool, right?

## Should I deploy this into production?

Probably not. But if you do, I'll buy you a beverage of your choice at AWS
re:Invent 2023.

[vpc-lattice]: https://aws.amazon.com/vpc/lattice/
