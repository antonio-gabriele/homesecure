import { Component, OnDestroy, OnInit } from '@angular/core';
import { MultiRouterService } from './framework/multirouter.service';
import { Subscription } from 'rxjs';

@Component({
  selector: 'mes',
  templateUrl: './mes.component.html'
})
export class MesComponent implements OnInit, OnDestroy {

  channels: any[] = [];
  subscription: Subscription | undefined;
  channel: any;
  properties: any[] = [];
  commands: any[] = [];

  ngOnDestroy(): void {
    this.subscription?.unsubscribe();
  }

  async ngOnInit(): Promise<void> {
    this.subscription = this.multiRouterService.Event.subscribe((e: any) => {
      switch (e.eventName) {
        case "channels/properties/status": {
          console.log(e.result);
          const channels = this.channels.filter(c => c.channel === e.result.channel);
          if (channels.length === 1) {
            const behaviours = channels[0].behaviours.filter((p: { behaviour: any; }) => p.behaviour === e.result.behaviour);
            if (behaviours.length === 1) {
              const properties = behaviours[0].properties.filter((p: { property: any; }) => p.property === e.result.property);
              if (properties.length === 1) {
                Object.assign(properties[0], e.result);
              }
            }
          }
        }
          break;
      }
    });
    await this.Channels();
  }

  constructor(private multiRouterService: MultiRouterService) {
  }

  async Test(): Promise<void> {
    var connected = await this.multiRouterService.Test();
    console.log(connected ? "Connected" : "Not Connected");
  }

  async Channels(): Promise<void> {
    this.channels = await this.multiRouterService.RPC("channels/channels", {}, true);
    console.log(this.channels);
  }

  async Command(behaviour: any, command: any): Promise<void> {
    await this.multiRouterService.RPC("channels/commands/execute", {
      behaviour: behaviour,
      command: command
    }, true);
  }

  async Save(): Promise<void> {
    var channels = this.channels.filter(c => c.$modified);
    if (channels.length) {
      await this.multiRouterService.RPC("channels/channels/save", channels, true);
    }
  }
}
