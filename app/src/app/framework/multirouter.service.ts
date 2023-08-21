import { Injectable } from '@angular/core';
import { Subject, firstValueFrom } from 'rxjs';
import { IMqttMessage, MqttService } from "ngx-mqtt";
import { ec } from "elliptic";
import { uint8ArrayToBase64, arrayBufferToBase64, base64ToUint8Array } from 'base64-u8array-arraybuffer'
import { gzip, ungzip } from 'pako'

class Pending {
  constructor(public Resolve: (value: any) => void, public Topic: string, public Plain: any) {
  }
}

@Injectable()
export class MultiRouterService {
  private Identifier: string = "a1b2c3";
  private OwnKeyPair: ec.KeyPair;
  private SharedKey: CryptoKey | null | undefined;
  private OwnPublicKey: string;
  private Promises: Map<string, Pending> = new Map<string, Pending>();
  public Event = new Subject();
  constructor(private _mqttService: MqttService) {
    var ec1 = new ec('p256');
    this.OwnKeyPair = ec1.genKeyPair();
    this.OwnPublicKey = uint8ArrayToBase64(this.OwnKeyPair.getPublic().encode('array', false));
    var topic = this.OwnPublicKey.replaceAll("/", "").replaceAll("+", "").replaceAll("=", "");
    this._mqttService.observe(`${this.Identifier}/${topic}/repair`).subscribe(m => this.Repair(m));
    this._mqttService.observe(`${this.Identifier}/${topic}/pair`).subscribe(m => this.Pair(m));
    this._mqttService.observe(`${this.Identifier}/${topic}/recv`).subscribe(m => this.Recv(m));
  }

  public async Test(): Promise<boolean> {
    let random = uint8ArrayToBase64(crypto.getRandomValues(new Uint8Array(12)));
    const result = await this.RPC(`edge/ping`, { ping: random }, false);
    return result.pong === random;
  }

  public async RPC(topic: string, plain: any, fireForget: boolean = false): Promise<any> {
    console.log('RPC 0');
    if (!this.SharedKey) {
      await this.RequestPair();
    }
    const iv = crypto.getRandomValues(new Uint8Array(12));
    const corrrelationId = uint8ArrayToBase64(crypto.getRandomValues(new Uint8Array(32)));
    const json = JSON.stringify({
      fireForget: fireForget,
      cid: corrrelationId,
      topic: topic,
      plain: plain
    });
    //console.log(`-> ${json}`);
    const encoded = new TextEncoder().encode(json);
    const compressed = gzip(encoded);
    crypto.subtle.encrypt({ "name": "AES-GCM", "iv": iv }, this.SharedKey!, compressed).then(cipher => {
      const payload = JSON.stringify({
        pk: this.OwnPublicKey,
        iv: uint8ArrayToBase64(iv),
        cipher: arrayBufferToBase64(cipher)
      });
      firstValueFrom(this._mqttService.publish(`${this.Identifier}/edge/recv`, payload)).then(() => {
        console.log('RPC 1');
      });
    });
    return new Promise<any>((resolve, _) => this.Promises.set(corrrelationId, new Pending(resolve, topic, plain)));
  }

  private Pair(mqttMessage: IMqttMessage) {
    //console.log('Pair <-');
    var ec1 = new ec('p256');
    var json = JSON.parse(mqttMessage.payload.toString());
    var pubKey1 = ec1.keyFromPublic(base64ToUint8Array(json.pk))
    var sharedKey = this.OwnKeyPair.derive(pubKey1.getPublic()).toArray();
    crypto.subtle.importKey('raw', new Uint8Array(sharedKey), {
      "name": "AES-GCM"
    }, false, ['encrypt', 'decrypt']).then(key => {
      this.SharedKey = key;
      this.Promises.get('pair')?.Resolve(true);
      //console.log('Pair ->');
    });
  }

  private Repair(_: IMqttMessage) {
    //console.log('Repair <-');
    this.RequestPair();
    //console.log('Repair ->');
  }

  private Recv(mqttMessage: IMqttMessage) {
    console.log('Recv 0');
    var json = JSON.parse(mqttMessage.payload.toString());
    var iv = base64ToUint8Array(json.iv);
    var cipher = base64ToUint8Array(json.cipher);
    crypto.subtle.decrypt({ "name": "AES-GCM", "iv": iv }, this.SharedKey!, cipher).then(plain => {
      const arrayBuffer = ungzip(plain);
      const plainText = new TextDecoder().decode(arrayBuffer);
      const json = JSON.parse(plainText);
      console.log(`Recv 1`);
      const promise = this.Promises.get(json.cid);
      if (promise) {
        promise.Resolve(json.plain);
      } else {
        this.Event.next(json.plain);
      }
    });
  }

  private async RequestPair(): Promise<void> {
    //console.log('RequestPair <-');
    var request = JSON.stringify({
      pk: this.OwnPublicKey
    });
    return new Promise<void>((resolve, _) => {
      this.Promises.set('pair', new Pending(resolve, "", null));
      firstValueFrom(this._mqttService.publish(`${this.Identifier}/edge/pair`, request));
      //console.log('RequestPair ->');
    });
  }
}