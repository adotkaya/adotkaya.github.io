---
title: "Back to School: Leaving Ericsson for an M.Sc. in Distributed Systems at Yaşar"
date: "2026-09-27"
slug: "back-to-school-msc-distributed-systems"
---

Seven months ago I left Ericsson.

No drama, no burnout email, no big announcement. I just knew I needed to move on, so I did.

## How I decided to move on

I joined Ericsson in September 2024 as a Software Developer. On paper it was good work: high-availability .NET Core backend services for a major telecom operator, 100K+ daily requests, SQL optimization, UAT and production deployments. I learned how corporate telecom software actually runs.

But day to day, most of the work was surface-level: CRUD endpoints, business logic on top of a database, process automation, testing pipelines. It paid well and it was stable. It just wasn't going anywhere I wanted to go.

At the same time, something obvious was happening outside. AI got good enough to do CRUD to database work end-to-end. And it's only getting better. I kept thinking: if I stay here doing surface-level development, what am I actually worth in two years?

That question made the decision for me.

I don't think the job market is cooked for engineers in general. I think it's cooked for surface-level development. The developer who only knows how to wire a form to a database is competing directly with an agent now. The developer who understands what's underneath: networking, concurrency, storage, consensus, failure modes and who knows how to orchestrate agents and approve a true implementation, that developer is more valuable than ever.

I wanted to be the second kind. Staying where I was wasn't going to get me there. So I left in early 2026.

## The 7 months in between

After I left, I gave myself one job: earn enough to buy uninterrupted time, then use that time properly.

I won't go into the finance details here, it was a means to an end, not the story. The point is it worked. It gave me a runway of a few months with no meetings, no tickets, no on-call.

I spent it studying and building.

I went deep into Go. I built proglog from *Distributed Services with Go*, not just typing it out, but tracing every layer until I actually understood it. I wrote about that process here. I kept working on dist_file_storage. I started reading *Designing Data-Intensive Applications* properly, not skimming. I wrote about Go concurrency patterns and system design on this blog.

That period clarified everything. Solo study can get you to passing tests and working demos. It can't easily give you depth, feedback, or research direction. I kept hitting questions I wanted to discuss with people who do this for a living. I wanted structure.

So I decided to go back to academia.

## Why Yaşar, why M.Sc., why now

I'm now pursuing an M.Sc. in Computer Engineering at Yaşar University, the same place I did my B.Sc.

Going back to the same university might look like going in circles. It does not feel that way to me.
Observing and learning from other universities' work and academic culture could have been a great addition to my thinking, I accept that. But to be honest I did not even get the chance to learn mine in B.Sc.

I was focused on work, open-source, startups, corporate... money for my whole university life, just graduating was okay for me because that was the dream that industry sold to us.

Economy, job market is booming. Just graduate, be an 'OK' developer and you will get a job and be happy. Rest will follow with experience. Well that was a mistake.

Had great professors in my B.Sc. and I did not focus on learning from them. Now it is time to fix that mistake.

For the first ~6 months, my focus is deliberately boring: distributed systems basics. Networking fundamentals, replication, consensus like Raft and Paxos, storage engines, fault tolerance. No AI hype, no agents, no frameworks. Just the fundamentals I skipped or half-learned while shipping business software.

## Where this goes: Distributed Systems + AI

After the basics are solid, I want to move toward Distributed Systems + AI.

My bet is simple: AI workloads are fundamentally distributed systems problems now. Training runs across hundreds of GPUs that fail constantly. Inference has to be served with low latency all over the world. LLM agents need memory, coordination, consistency, all the old hard problems, back with new names.

That's where I think the interesting research and the real work options will be in the next few years. Not in writing CRUD wrappers around an LLM API, but in building the infrastructure that makes large-scale AI reliable, observable, and efficient.

I'm keeping this intentionally high-level. I don't have a thesis proposal or a paper list yet. First the foundations, then the intersection. I'll write about it here as it sharpens.

## What this blog becomes

This site started as a portfolio. Then it became a place for technical deep-dives. Now it'll also be my public notebook for the M.Sc.

Expect more notes on distributed systems fundamentals, follow-ups to the proglog series, and eventually experiments at the DistSys + AI boundary. Same style as before: what I built, what broke, what I actually learned.

Seven months ago I left a stable job because surface-level work felt like a dead end. Now I'm back at school to learn the deep stuff properly.

Feels like the right time to bet, also the last exit before the tunnel.

Hoping that 'a moving man, will eventually find his luck'.
