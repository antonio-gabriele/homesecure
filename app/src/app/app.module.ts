import { NgModule } from '@angular/core';
import { BrowserModule } from '@angular/platform-browser';
import { AppRoutingModule } from './app-routing.module';
import { MesComponent } from './mes.component';
import { HttpClientModule } from '@angular/common/http';
import { FormsModule } from '@angular/forms';
import { AppComponent } from './app.component';
import { MqttModule, IMqttServiceOptions } from "ngx-mqtt";
import { GeneralComponent } from './general.component';
import { BeeComponent } from './bee.component';
import { NgModelChangeDebouncedDirective } from './framework/debounce.directive';
import { Bee1Component } from './bee1.component';
import { BrowserAnimationsModule } from '@angular/platform-browser/animations';

import { MatToolbarModule } from '@angular/material/toolbar';
import { MatCardModule } from '@angular/material/card';
import { MatExpansionModule } from '@angular/material/expansion';
import { MatIconModule } from '@angular/material/icon';
import { MatRippleModule } from '@angular/material/core';
import { MatButtonModule } from '@angular/material/button';
import { MatSlideToggleModule } from '@angular/material/slide-toggle';
import { MatFormFieldModule } from '@angular/material/form-field';
import { MatInputModule } from '@angular/material/input';
import { MatListModule } from '@angular/material/list';
import { MatDividerModule } from '@angular/material/divider';
import { MatGridListModule } from '@angular/material/grid-list';
import { FlexLayoutModule } from '@angular/flex-layout';
import { MatSliderModule } from '@angular/material/slider';

export const MQTT_SERVICE_OPTIONS_WAN: IMqttServiceOptions = {
  hostname: 'homesecure.dev',
  protocol: 'wss',
  username: 'user',
  password: 'bitnami',
  port: 15676,
  path: '/ws'
}
export const MQTT_SERVICE_OPTIONS_LAN: IMqttServiceOptions = {
  hostname: '192.168.1.211',
  protocol: 'wss',
  port: 8443,
  path: '/',
}
@NgModule({
  declarations: [
    AppComponent,
    MesComponent,
    GeneralComponent,
    BeeComponent,
    Bee1Component,
    NgModelChangeDebouncedDirective
  ],
  imports: [
    BrowserModule,
    FlexLayoutModule,
    MatGridListModule,
    MatSliderModule,
    MatSlideToggleModule,
    MatButtonModule,
    MatListModule,
    MatToolbarModule,
    MatCardModule,
    MatExpansionModule,
    MatIconModule,
    MatFormFieldModule,
    MatInputModule,
    MatDividerModule,
    BrowserAnimationsModule,
    MatRippleModule,
    AppRoutingModule,
    HttpClientModule,
    FormsModule,
    MqttModule.forRoot(MQTT_SERVICE_OPTIONS_WAN)
  ],
  bootstrap: [AppComponent]
})
export class AppModule { }
