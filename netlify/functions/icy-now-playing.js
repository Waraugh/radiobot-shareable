// Experimental, single-station ICY metadata probe. Never accepts arbitrary URLs.
const STREAMS=['https://streams.90s90s.de/bawue/mp3-192/streams.90s90s.de/','https://streams.90s90s.de/bawue/mp3-128/streams.90s90s.de/'];
let cached=null, cachedAt=0;
exports.handler=async()=>{
  const headers={'Content-Type':'application/json','Cache-Control':'no-store'};
  if(cached && Date.now()-cachedAt<15000)return {statusCode:200,headers,body:JSON.stringify(cached)};
  const controller=new AbortController();
  const timeout=setTimeout(()=>controller.abort(),7000);
  try{
    let response, failures=[];
    for(const url of STREAMS){
      try{response=await fetch(url,{headers:{'Icy-MetaData':'1','User-Agent':'RicanRadioMetadataTest/1.0'},signal:controller.signal});break;}
      catch(e){failures.push({endpoint:url.includes('mp3-192')?'192':'128',message:e.message,cause:e.cause?.code||e.cause?.message||'unknown'});}
    }
    if(!response)return {statusCode:503,headers,body:JSON.stringify({error:'Stream connection failed',failures})};
    if(!response.ok)throw new Error('Stream HTTP '+response.status);
    const interval=Number(response.headers.get('icy-metaint'));
    if(!Number.isInteger(interval)||interval<1||interval>1048576)throw new Error('No ICY metadata interval');
    const reader=response.body.getReader();
    let buffer=new Uint8Array(0);
    const needed=interval+1;
    while(buffer.length<needed){
      const {value,done}=await reader.read();
      if(done)throw new Error('Stream ended before metadata');
      const merged=new Uint8Array(buffer.length+value.length);
      merged.set(buffer);merged.set(value,buffer.length);buffer=merged;
    }
    const length=buffer[interval]*16;
    if(length>4080)throw new Error('Invalid ICY block length');
    while(buffer.length<needed+length){
      const {value,done}=await reader.read();
      if(done)throw new Error('Stream ended during metadata');
      const merged=new Uint8Array(buffer.length+value.length);
      merged.set(buffer);merged.set(value,buffer.length);buffer=merged;
    }
    const raw=new TextDecoder('utf-8').decode(buffer.subarray(needed,needed+length)).replace(/\0/g,'');
    const title=raw.match(/StreamTitle='([^']*)'/i)?.[1]?.trim()||'';
    if(!title)throw new Error('Stream does not expose a song title');
    const parts=title.split(/\s+-\s+/);
    if(parts.length<2)throw new Error('Unrecognized title: '+title.slice(0,100));
    const artist=parts.shift().trim(),song=parts.join(' - ').trim();
    cached={artist,song,raw:title};cachedAt=Date.now();
    return {statusCode:200,headers,body:JSON.stringify(cached)};
  }catch(error){
    return {statusCode:503,headers,body:JSON.stringify({error:String(error.message||error),cause:error.cause?.code||error.cause?.message||null})};
  }finally{clearTimeout(timeout);controller.abort();}
};