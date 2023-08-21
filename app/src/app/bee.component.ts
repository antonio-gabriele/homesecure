import { Component, OnDestroy, OnInit } from '@angular/core';
import { MultiRouterService } from './framework/multirouter.service';
import { Subscription } from 'rxjs';
import { end } from '@popperjs/core';

@Component({
  selector: 'app-bee',
  templateUrl: './bee.component.html'
})
export class BeeComponent implements OnInit, OnDestroy {
  channels: any[] = [];
  behaviours: any[] = [];
  permitJoin: boolean = false;
  subscription: Subscription | undefined;
  constructor(private multiRouterService: MultiRouterService) {
  }

  ngOnDestroy(): void {
    this.subscription?.unsubscribe();
  }

  async ngOnInit(): Promise<void> {
    this.subscription = this.multiRouterService.Event.subscribe((e: any) => {
      console.log(e);
      if (e.eventName === 'bee/permitJoin') {
        //this.channels = this.channels.concat(e.result);
      }
      if (e.eventName === 'channels/behaviours') {
        this.behaviours = this.behaviours.concat(e.result);
      }
    });
    await this.Channels();
  }

  async PermitJoin(): Promise<void> {
    var result = await this.multiRouterService.RPC("bee/permitJoin", {}, true);
    console.log(result);
  }

  async ReadBindingTable(channel: any): Promise<void> {
    var result = await this.multiRouterService.RPC("bee/readBindingTable", channel, true);
    console.log(result);
  }

  async Explore(): Promise<void> {
    var result = await this.multiRouterService.RPC("bee/explore", {}, true);
    console.log(result);
  }

  async Channels(): Promise<void> {
    this.channels = await this.multiRouterService.RPC("channels/channels/items", {});
    console.log(this.channels);
  }

  async Save(endpoint: string, obj: any) {
    const transferObject = {};
    Object.assign(transferObject, obj);
    this.Clean(transferObject);
    await this.multiRouterService.RPC(endpoint, transferObject, true);
  }

  async BehaviourSave(behaviour: any): Promise<void> {
    await this.Save("channels/behaviours/wr", behaviour);
  }

  async ChannelSave(channel: any): Promise<void> {
    await this.Save("channels/channels/wr", channel);
  }

  Clean(object: any) {
    const properties = Object.getOwnPropertyNames(object);
    for (const property of properties) {
      if (typeof (object[property]) === typeof ([])) {
        delete (object[property]);
      }
      if (typeof (object[property]) === typeof ({})) {
        delete (object[property]);
      }
    }
  }
}
