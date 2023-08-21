import { Component, OnDestroy, OnInit } from '@angular/core';
import { MultiRouterService } from './framework/multirouter.service';
import { ActivatedRoute } from '@angular/router';

@Component({
  selector: 'app-bee1',
  templateUrl: './bee1.component.html'
})
export class Bee1Component implements OnInit, OnDestroy {
  channels: any[] = [];
  behaviours: any[] = [];
  permitJoin: boolean = false;
  subscription: any;
  channel: string | undefined;
  behaviour: any;
  commands: any;
  properties: any;
  constructor( //
    private activatedRoute: ActivatedRoute, //
    private multiRouterService: MultiRouterService) {
  }

  ngOnInit() {
    this.subscription = this.activatedRoute.params.subscribe(async params => {
      this.channel = params['channel'];
      this.behaviours = await this.multiRouterService.RPC("bee/behaviours", {
        channel: this.channel
      });
    });
  }

  ngOnDestroy() {
    this.subscription.unsubscribe();
  }

  async Refresh() {
    this.commands = [];
    this.properties = [];
    if (this.behaviour.enabled) {
      this.commands = await this.multiRouterService.RPC("bee/commands", {
        channel: this.channel,
        behaviour: this.behaviour.behaviour
      });
      this.properties = await this.multiRouterService.RPC("bee/properties", {
        channel: this.channel,
        behaviour: this.behaviour.behaviour
      });
    }
    console.log(this.commands)
  }

  async BehaviourToggle(behaviour: any): Promise<void> {
    behaviour.enabled = !behaviour.enabled;
    await this.multiRouterService.RPC("behaviours/save", [behaviour], true);
    await this.Refresh();
  }

  async CommandToggle(command: any): Promise<void> {
    command.enabled = !command.enabled;
    await this.multiRouterService.RPC("commands/save", [command], true);
  }

  async PropertyToggle(property: any): Promise<void> {
    property.enabled = !property.enabled;
    await this.multiRouterService.RPC("properties/save", [property], true);
  }

  async Open(content: any, behaviour: any) {
    this.behaviour = behaviour;
    await this.Refresh();
    //this.offcanvasService.open(content, { ariaLabelledBy: 'offcanvas-basic-title' });
  }
}
