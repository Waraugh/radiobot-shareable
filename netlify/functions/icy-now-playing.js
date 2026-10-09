// Experimental 90s Hits ICY reader. Fixed upstream URLs; no user-provided proxy target.
const https=require('node:https');
const STREAMS=['https://streams.90s90s.de/bawue/mp3-192/streams.90s90s.de/','https://streams.90s90s.de/bawue/mp3-128/streams.90s90s.de/'];
let cached=null,cachedAt=0;
function probe(url,redirects=0){
 return new Promise((resolve,reject)=>{
  const request=https.get(url,{insecureHTTPParser:true,headers:{'Icy-MetaData':'1','User-Agent':'RicanRadioMetadataTest/1.0'},timeout:6500},response=>{
   const code=response.statusCode||0;
   if(code>=300&&code<400&&response.headers.location){
    response.destroy();
    if(redirects>=3)return reject(new Error('Too many redirects'));
    const next=new URL(response.headers.location,url);
    if(next.protocol!=='https:'||!/(^|\\.)(90s90s\\.de|radiobob\\.de|radio\\.de|streamabc\\.net|regiocast\\.de|laut\\.fm)$/.test(next.hostname))return reject(new Error('Unexpected redirect host: '+next.hostname));
    return resolve(probe(next.toString(),redirects+1));
   }
   if(code!==200){response.destroy();return reject(new Error('HTTP '+code));}
   const interval=Number(response.headers['icy-metaint']);
   if(!Number.isInteger(interval)||interval<1||interval>1048576){response.destroy();return reject(new Error('No ICY metadata interval'));}
   let chunks=[],length=0,needed=interval+1,metaLength=null;
   response.on('data',chunk=>{
    chunks.push(chunk);length+=chunk.length;
    if(length<needed)return;
    let buf=Buffer.concat(chunks,length);
    if(metaLength===null){
     metaLength=buf[interval]*16;
     if(metaLength>4080){response.destroy();return reject(new Error('Invalid ICY block length'));}
     needed=interval+1+metaLength;
    }
    if(length<needed)return;
    const raw=buf.subarray(interval+1,needed).toString('utf8').replace(/\\0/g,'');
    response.destroy();
    const title=raw.match(/StreamTitle='([^']*)'/i)?.[1]?.trim()||'';
    if(!title)return reject(new Error('Stream does not expose a song title'));
    const parts=title.split(/\\s+-\\s+/);
    if(parts.length<2)return reject(new Error('Unrecognized title: '+title.slice(0,100)));
    resolve({artist:parts.shift().trim(),song:parts.join(' - ').trim(),raw:title});
   });
   response.on('end',()=>reject(new Error('Stream ended before metadata')));
   response.on('error',reject);
  });
  request.on('timeout',()=>request.destroy(new Error('Connection timed out')));
  request.on('error',reject);
 });
}
exports.handler=async()=>{
 const headers={'Content-Type':'application/json','Cache-Control':'no-store'};
 if(cached&&Date.now()-cachedAt<15000)return {statusCode:200,headers,body:JSON.stringify(cached)};
 const failures=[];
 for(const url of STREAMS){
  try{
   const result=await probe(url);
   cached=result;cachedAt=Date.now();
   return {statusCode:200,headers,body:JSON.stringify(result)};
  }catch(e){failures.push({endpoint:url.includes('mp3-192')?'192':'128',error:e.message,code:e.code||null});}
 }
 return {statusCode:503,headers,body:JSON.stringify({error:'ICY probe failed',failures})};
};
